package tftp

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Server TFTP服务器
type Server struct {
	addr      string
	conn      *net.UDPConn
	filesRoot string
}

// NewServer 创建TFTP服务器
func NewServer(addr string, filesRoot string) *Server {
	return &Server{
		addr:      addr,
		filesRoot: filesRoot,
	}
}

// Start 启动TFTP服务器
func (s *Server) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen UDP: %w", err)
	}

	s.conn = conn
	log.Printf("✅ TFTP服务器启动成功: %s", s.addr)
	log.Printf("📁 文件根目录: %s", s.filesRoot)

	// 启动请求处理循环
	go s.serve()

	return nil
}

// Stop 停止TFTP服务器
func (s *Server) Stop() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// serve 处理TFTP请求
func (s *Server) serve() {
	buffer := make([]byte, 516) // TFTP最大包大小

	for {
		n, clientAddr, err := s.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("⚠️  TFTP读取错误: %v", err)
			continue
		}

		// 异步处理每个请求
		go s.handleRequest(buffer[:n], clientAddr)
	}
}

// handleRequest 处理单个TFTP请求
func (s *Server) handleRequest(data []byte, clientAddr *net.UDPAddr) {
	if len(data) < 4 {
		return
	}

	opcode := uint16(data[0])<<8 | uint16(data[1])

	switch opcode {
	case 1: // RRQ (Read Request)
		s.handleReadRequest(data[2:], clientAddr)
	case 2: // WRQ (Write Request)
		s.sendError(clientAddr, 2, "Write not supported")
	default:
		s.sendError(clientAddr, 4, "Illegal TFTP operation")
	}
}

// handleReadRequest 处理读取请求
func (s *Server) handleReadRequest(data []byte, clientAddr *net.UDPAddr) {
	// 解析文件名
	filename, mode := s.parseRRQ(data)
	if filename == "" {
		s.sendError(clientAddr, 4, "Invalid filename")
		return
	}

	log.Printf("📥 TFTP RRQ: %s (mode: %s) from %s", filename, mode, clientAddr)

	// 打开文件
	filePath := filepath.Join(s.filesRoot, filename)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("❌ 文件不存在: %s", filePath)
		s.sendError(clientAddr, 1, "File not found")
		return
	}
	defer file.Close()

	// 发送文件数据
	s.sendFile(file, clientAddr)
}

// sendFile 发送文件数据
func (s *Server) sendFile(file *os.File, clientAddr *net.UDPAddr) {
	// 创建新的UDP连接用于数据传输
	conn, err := net.DialUDP("udp", nil, clientAddr)
	if err != nil {
		log.Printf("❌ 无法连接客户端: %v", err)
		return
	}
	defer conn.Close()

	blockNum := uint16(1)
	buffer := make([]byte, 512)

	for {
		// 读取数据块
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			log.Printf("❌ 读取文件错误: %v", err)
			return
		}

		// 构建DATA包
		dataPacket := make([]byte, 4+n)
		dataPacket[0] = 0x00 // Opcode: DATA (高字节)
		dataPacket[1] = 0x03 // Opcode: DATA (低字节)
		dataPacket[2] = byte(blockNum >> 8)
		dataPacket[3] = byte(blockNum & 0xFF)
		copy(dataPacket[4:], buffer[:n])

		// 发送数据包（带重试）
		ackReceived := false
		for retry := 0; retry < 5; retry++ {
			_, writeErr := conn.Write(dataPacket)
			if writeErr != nil {
				log.Printf("⚠️  发送数据失败: %v", writeErr)
				continue
			}

			// 等待ACK
			conn.SetReadDeadline(time.Now().Add(3 * time.Second))
			ackBuffer := make([]byte, 516)
			ackN, readErr := conn.Read(ackBuffer)

			if readErr == nil && ackN >= 4 {
				ackOpcode := uint16(ackBuffer[0])<<8 | uint16(ackBuffer[1])
				ackBlock := uint16(ackBuffer[2])<<8 | uint16(ackBuffer[3])

				if ackOpcode == 4 && ackBlock == blockNum {
					ackReceived = true
					break
				}
			}

			time.Sleep(100 * time.Millisecond)
		}

		if !ackReceived {
			log.Printf("❌ 未收到ACK for block %d", blockNum)
			return
		}

		// 最后一个数据包（小于512字节）
		if n < 512 {
			log.Printf("✅ 文件传输完成: %d blocks", blockNum)
			return
		}

		blockNum++
	}
}

// sendError 发送错误响应
func (s *Server) sendError(clientAddr *net.UDPAddr, errorCode uint16, errorMsg string) {
	packet := make([]byte, 5+len(errorMsg))
	packet[0] = 0x00 // Opcode: ERROR (高字节)
	packet[1] = 0x05 // Opcode: ERROR (低字节)
	packet[2] = byte(errorCode >> 8)
	packet[3] = byte(errorCode & 0xFF)
	copy(packet[4:], errorMsg)
	packet[len(packet)-1] = 0x00 // 结束符

	s.conn.WriteToUDP(packet, clientAddr)
}

// parseRRQ 解析RRQ包
func (s *Server) parseRRQ(data []byte) (filename string, mode string) {
	// RRQ格式: 文件名\0模式\0
	parts := make([]string, 0, 2)
	start := 0

	for i, b := range data {
		if b == 0 {
			parts = append(parts, string(data[start:i]))
			start = i + 1

			if len(parts) >= 2 {
				break
			}
		}
	}

	if len(parts) >= 2 {
		return parts[0], parts[1]
	}

	return "", ""
}
