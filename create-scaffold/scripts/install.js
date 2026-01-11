#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const https = require("https");
const { execSync } = require("child_process");

const VERSION = "0.1.0";
const BINARY_NAME = "scaffold";
const GITHUB_REPO = "makemore/scaffold";

// Map Node.js platform/arch to Go platform/arch
const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

function getPlatform() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    console.error(`Unsupported platform: ${process.platform}/${process.arch}`);
    process.exit(1);
  }

  return { platform, arch };
}

function getBinaryPath() {
  const binDir = path.join(__dirname, "..", "bin");
  const ext = process.platform === "win32" ? ".exe" : "";
  return path.join(binDir, `${BINARY_NAME}${ext}`);
}

function getDownloadUrl() {
  const { platform, arch } = getPlatform();
  const ext = platform === "windows" ? ".exe" : "";
  return `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/scaffold-${platform}-${arch}${ext}`;
}

async function downloadBinary(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);

    const request = (url) => {
      https
        .get(url, (response) => {
          // Handle redirects
          if (response.statusCode === 302 || response.statusCode === 301) {
            request(response.headers.location);
            return;
          }

          if (response.statusCode !== 200) {
            reject(new Error(`Failed to download: ${response.statusCode}`));
            return;
          }

          response.pipe(file);
          file.on("finish", () => {
            file.close();
            resolve();
          });
        })
        .on("error", (err) => {
          fs.unlink(dest, () => {});
          reject(err);
        });
    };

    request(url);
  });
}

async function install() {
  const binaryPath = getBinaryPath();
  const binDir = path.dirname(binaryPath);

  // Create bin directory if it doesn't exist
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  // Check if binary already exists
  if (fs.existsSync(binaryPath)) {
    console.log("Scaffold binary already installed.");
    return;
  }

  const downloadUrl = getDownloadUrl();
  console.log(`Downloading scaffold from ${downloadUrl}...`);

  try {
    await downloadBinary(downloadUrl, binaryPath);

    // Make binary executable on Unix
    if (process.platform !== "win32") {
      fs.chmodSync(binaryPath, 0o755);
    }

    console.log("Scaffold installed successfully!");
  } catch (error) {
    console.error(`Failed to download scaffold: ${error.message}`);
    console.error(
      "\nYou can manually download the binary from:"
    );
    console.error(
      `https://github.com/${GITHUB_REPO}/releases`
    );
    process.exit(1);
  }
}

install();

