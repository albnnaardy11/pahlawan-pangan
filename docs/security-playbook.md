---
id: security-playbook
title: Security Playbook: Argon2id & RS256
---

# 🛡️ Pahlawan Pangan Security Architecture

## 1. Password Hashing: Argon2id
We do not use BCrypt. We use **Argon2id**, the winner of the Password Hashing Competition (PHC).

### Configuration
*   **Time**: 1 (Iterations)
*   **Memory**: 64MB (Hardness)
*   **Threads**: 4 (Parallelism)
*   **Salt**: 16 bytes (Cryptographically secure random)

### Why?
Argon2id is hybrid. It resists:
1.  **Side-Channel Attacks** (like Argon2i)
2.  **GPU Cracking Attacks** (like Argon2d)

## 2. Token Signing: RS256
We utilize **Asymmetric Keys** (RSA SHA-256).

### Flow
1.  **Auth Service** holds the `private.pem`. Only it can sign tokens.
2.  **Downstream Services** (Logistics, Notification) hold `public.pem`. They can *only* verify.

### Attack Vector Mitigation
*   **Key Leakage**: If a downstream service is compromised, the attacker only gets the Public Key. They simply cannot forge new Admin tokens.
*   **Algorithm Confusion**: We strictly enforce `if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok` to prevent "None" algo attacks.
