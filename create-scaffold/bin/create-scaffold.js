#!/usr/bin/env node

/**
 * create-scaffold - npx wrapper for scaffold CLI
 * 
 * Usage:
 *   npx create-scaffold myapp
 *   npx create-scaffold myapp --base django
 *   npx create-scaffold --help
 * 
 * This is equivalent to:
 *   scaffold init myapp
 *   scaffold init myapp --base django
 *   scaffold init --help
 */

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

  // Prepend "init" to arguments for create-scaffold usage
  // npx create-scaffold myapp -> scaffold init myapp
  const userArgs = process.argv.slice(2);
  
  // If first arg is --help or -h, show init help
  // If first arg is --version or -v, show version
  const firstArg = userArgs[0];
  let args;
  
  if (firstArg === "--version" || firstArg === "-v") {
    args = ["version"];
  } else if (firstArg === "--help" || firstArg === "-h" || !firstArg) {
    args = ["init", "--help"];
  } else {
    args = ["init", ...userArgs];
  }

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

