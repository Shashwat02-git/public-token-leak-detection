# 🛡️ Public Token Leakage Detection System

A **high-performance, concurrent Go service** that scans for sensitive tokens, enriches the findings with live geolocation data, and sends immediate, detailed alerts via **Slack** and **Email**.

This project is a **fully-containerized, cloud-deployable server** simulating a real-world leakage detection pipeline.

---

## 🚀 Key Features

### ⚙️ High-Performance Concurrency
- Built with **Go worker pools** and **concurrent-safe caching (`sync.Map`)**.
- Scans and processes files efficiently across multiple goroutines.

### 🧪 Simulation-Based Scanning
- Scans a local directory `/source_files` with mock metadata (`metadata.json`).
- Simulates real leaks from sources like GitHub for **repeatable testing**.

### 🌍 Live Geolocation Enrichment
- Extracts the committer's (mock) IP address from metadata.
- Uses **[ip-api.com](https://ip-api.com)** to enrich alerts with real-time geolocation data.

### 📢 Rich Dual-Channel Alerting
- Sends alerts to:
  - **Slack** (using formatted Blocks)
  - **Email** (using custom HTML templates)

### 🧭 Actionable Alerts
Each alert includes:
- Token type  
- Owner  
- File path  
- Geolocation info  
- Remediation steps  

### ☁️ Cloud-Ready API
- **Endpoints:**
  - `/` — Health check  
  - `/api/check` — Manually trigger a scan

### 🐳 Fully Dockerized
- **Multi-stage Dockerfile** for small, secure production builds.
- Portable, reproducible, and easy to deploy.

---

## 🌐 Live Demo & Demonstration

**Hosted on Render**

- **Live URL:**  
  `https://public-token-leak-detection.onrender.com`

- **Health Check:**  
  `https://public-token-leak-detection.onrender.com/`

- **Trigger Scan:**  
  `https://public-token-leak-detection.onrender.com/api/check`

---

## 🔍 Demonstration: Finding a Planted Token

To demonstrate the scanner’s capabilities, a token from `inventory.json` was intentionally planted in a simulated source file.

### The Inventory
`inventory.json` lists an AWS Access Key ID: AKIAJVW6QEXAMPLE7FHFQ belonging to **shashwatbhargava2@gmail.com**

### The Public Leak
This exact token was added to `source_files/app1/app.py`

### The Metadata
Simulated  metadata stored in `source_files/app1/metadata.json` with an `authorIP` of **8.8.8.8**

### The Scan
When triggered:
1. The system finds the token.
2. Reads the commit metadata.
3. Enriches IP info (→ *Mountain View, California, USA*).
4. Sends detailed alerts.

---

## 📸 Evidence: Generated Alerts

### ✅ Slack Alert
*(Works on both local and deployed versions)*

**Screenshot:**
![Slack notification screenshot](./img/slack.png)

### 📧 Email Alert
*(Screenshot from local run — works locally, not on Render free-tier)*

**Screenshot:**
![Email notification screenshot](./img/email.png)

## ⚠️ Important Note on Email Deployment

This project illustrates a **real-world deployment challenge** with free-tier cloud providers.

### ✅ Slack Alerts
- Fully functional on the **live Render** deployment.

### ❌ Email (SMTP) Alerts
- Work locally ✅  
- Fail on Render ❌

### The Reason
Render (and similar platforms) block all outbound SMTP ports:
- **25**, **465**, **587**

**Error in logs:**
```
Email Error: Failed to send email: dial tcp 142.251.188.108:587: connect: connection timed out
```
- This confirms the **firewall restriction**, not a code issue.

---

## 🧰 Getting Started (Local Setup)

### 📋 Prerequisites
- Go **1.25.3+**
- A valid Gmail account with an **App Password**
- A **Slack Webhook URL**

---

### 🧩 Installation & Running

- Clone the repository:
```bash
git clone https://github.com/Shashwat02-git/public-token-leak-detection.git
cd public-token-leak-detection
```

- Copy and edit the environment file:
```
cp .env.sample .env
```

- Edit .env:
```
SENDER="your-gmail-address@gmail.com"
PASSWORD="your-google-app-password"
SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
```

- Run the application:
```
go run .
```

The server starts at:
```
http://localhost:8080
```

- Trigger a scan:
```
http://localhost:8080/api/check
```

### 🐳 Running with Docker
- Build the docker image:
```
docker build -t token-scanner .
```

- Run the docker container:
```
docker run --env-file .env -p 8080:8080 token-scanner

```

- Access the server:
```
http://localhost:8080
```






