#!/usr/bin/env node

const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");

const BINARY_NAME = "scaffold";

function getBinaryPath() {
  const ext = process.platform === "win32" ? ".exe" : "";
  return path.join(__dirname, `${BINARY_NAME}${ext}`);
}

function main() {
  const binaryPath = getBinaryPath();

  if (!fs.existsSync(binaryPath)) {
    console.error("Scaffold binary not found. Running install...");
    require("../scripts/install.js");
    
    // Check again after install
    if (!fs.existsSync(binaryPath)) {
      console.error("Failed to install scaffold binary.");
      process.exit(1);
    }
  }

  // Forward all arguments to the binary
  const args = process.argv.slice(2);
  
  const child = spawn(binaryPath, args, {
    stdio: "inherit",
    env: process.env,
  });

  child.on("error", (err) => {
    console.error(`Failed to run scaffold: ${err.message}`);
    process.exit(1);
  });

  child.on("exit", (code) => {
    process.exit(code || 0);
  });
}

main();

