# 🛰️ ClipSync
**End-to-End Encrypted Real-time Clipboard Sync.**

ClipSync is a minimalist, high-integrity utility for syncing text across devices. It uses a "Zero-Knowledge" architecture: data is encrypted in the browser before ever touching the wire.

### 🛡️ Security Architecture
- **AES-GCM Encryption:** Handled via Web Crypto API.
- **Fragment-Based Keys:** Encryption keys stay in the URL hash (`#`), ensuring they are never transmitted to the backend.
- **Blind Relay:** The Go backend acts as a high-speed traffic controller, routing encrypted blobs without knowing their contents.

### 🛠️ Tech Stack
- **Frontend:** SvelteKit, TailwindCSS, Lucide Icons.
- **Backend:** Go (Golang), Gorilla WebSockets, SQLite.
- **Environment:** Nix Flakes (for reproducible builds).

---

### 🚀 Local Development

#### 1. Enter the Environment
If you have Nix installed, the shell will provide Go, Bun, and SQLite automatically.
```bash
nix develop
```

#### 2. Backend Setup
```bash
cd backend
make run
```

#### 3. Frontend Setup
```bash
cd frontend
bun install
bun dev
```

### Deployment
#### Updating soon

### Roadmap
- [] TTL Worker: Auto delete snippets from SQLite after 24hrs
- [] E2E Testing: Automated playwright flows for cross-device verifications
- [] Argon2 for optional password derived keys
