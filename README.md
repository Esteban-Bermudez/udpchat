<html>
  <h1 align="center">udpchat</h1>
</html>

<p align="center">
  <!-- Replace these URLs with your actual badge image URLs -->
  <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/Esteban-Bermudez/udpchat">
  <img alt="Release" src="https://img.shields.io/github/v/release/Esteban-Bermudez/udpchat">
</p>

<p align="center">
  <!-- Add your animated GIF here -->
  <img alt="udp‑chat demo" src="path/to/demo.gif">
</p>

### 💡 Description

`udpchat` is a simple command-line application that helps peers connect via UDP using NAT traversal.

It works by:

1. Connecting to a [STUN](https://en.wikipedia.org/wiki/STUN) server to determine your **public IP and port**.
2. Prompting you to **share this address** with your peer.
3. Using **UDP hole punching** to establish a direct connection with the other peer.
4. Exchanging chat messages over that UDP connection.

---

### 📦 Installation

You can install the application in one of the following ways:

* **Using Go:**

  ```bash
  go install github.com/Esteban-Bermudez/udpchat/cmd/udpchat@v0.1.0
  ```

* **Manual build:**

  ```bash
  git clone https://github.com/Esteban-Bermudez/udpchat.git
  cd udpchat/cmd/udpchat
  go build -o udpchat
  ```

* **Pre-built binary:**

  [Download from the Releases section →](https://github.com/Esteban-Bermudez/udpchat/releases/tag/v0.1.0)

---

### 🚀 Usage

Once installed or built, simply run:

```bash
./udpchat
```

You will be prompted to:

* Connect to the STUN server and retrieve your public address.
* Share that address with your peer.
* Enter your peer’s address when prompted.

Once both peers have exchanged addresses, the chat session will begin.
