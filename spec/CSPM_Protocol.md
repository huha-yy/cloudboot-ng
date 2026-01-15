# CSPM_Protocol.md (CloudBoot Server Provider Mechanism)

## 1. Interaction Model (交互模型)
Providers are independent binaries executed by the `cb-exec` tool inside BootOS.
- **Communication**: JSON over Stdin (Input) and Stdout (Output).
- **Logs**: Structured JSON over Stderr.

## 2. Command Interface (CLI 规范)
Every Provider MUST implement these subcommands:

### 2.1 `probe`
- **Goal**: Check if this provider supports the current hardware.
- **Input**: None.
- **Output**: JSON containing supported hardware IDs found on the system.

### 2.2 `plan` (Dry Run)
- **Goal**: Calculate changes without applying them.
- **Input (Stdin)**:
  
  {
    "action": "plan",
    "resource": "raid",
    "desired_state": { "level": "10", "drives": "all" },
    "current_state": { ... } // Optional
  }


+ **Output (Stdout)**:

{
  "status": "success",
  "changes_required": true,
  "plan_summary": "Will delete VD 0 and create new VD 1 (RAID10)."
}


### 2.3 `apply` (Execution)
+ **Goal**: Make changes to hardware.
+ **Input (Stdin)**: Same as `plan`, but action is "apply".
+ **Output (Stdout)**:


{
  "status": "success",
  "data": { "created_resources": ["vd_1"] }
}


## 3. Security & DRM Protocol (安全与版权)
### 3.1 The Artifact: `.cbp` (CloudBoot Package)
A standard ZIP file containing:

+ `manifest.json`: Version and dependencies.
+ `watermark.json`: Download trace info (User ID, Time, Sig).
+ `provider.enc`: **AES-256 Encrypted** binary.

### 3.2 Decryption Flow (The "Key Separation")
1. **At Store (Cloud)**: Binary is encrypted with a global Master Key.
2. **At Core (On-Prem)**: 
    - User imports `.cbp`. Core verifies `watermark.json` signature.
    - Core uses Customer License Key to decrypt Master Key (Logic: Envelope Encryption).
    - Core re-encrypts the binary with a generated **Session Key** before sending to BootOS.
3. **At BootOS (RAM)**:
    - `cb-fetch` receives `provider.enc` + `Session Key`.
    - Decrypts to `/dev/shm/provider` (Memory only).
    - Executes and deletes immediately after exit.

## 4. User Overlay (微调机制)
To handle non-standard hardware behavior (The "Quirks"), CSPM supports configuration injection.

+ **Mechanism**: The `apply` input JSON includes an `overlay` field.
+ **Example**:

{
  "action": "apply",
  "params": { ... },
  "overlay": {
    "quirks": {
      "init_timeout_sec": 600,  // Override default 60s
      "ignore_battery": true    // Override default safety check
    }
  }
}


+ **Requirement**: All Providers MUST check the `overlay` field and override internal defaults if keys match.
