#!/usr/bin/env node
"use strict";

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const https = require("https");

const VERSION = require("./package.json").version;
const REPO = "tomenden/env-pilot";

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
    console.error(
      `Unsupported platform: ${process.platform}/${process.arch}`
    );
    process.exit(1);
  }

  return { platform, arch };
}

function getDownloadUrl() {
  const { platform, arch } = getPlatform();
  const ext = platform === "windows" ? "zip" : "tar.gz";
  const name = `env-pilot_${platform}_${arch}.${ext}`;
  return `https://github.com/${REPO}/releases/download/v${VERSION}/${name}`;
}

function download(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode === 302 || res.statusCode === 301) {
          return download(res.headers.location).then(resolve, reject);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`Download failed: HTTP ${res.statusCode}`));
        }
        const chunks = [];
        res.on("data", (chunk) => chunks.push(chunk));
        res.on("end", () => resolve(Buffer.concat(chunks)));
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

async function main() {
  const url = getDownloadUrl();
  const binDir = path.join(__dirname, "bin");
  const binName =
    process.platform === "win32" ? "env-pilot.exe" : "env-pilot";
  const binPath = path.join(binDir, binName);

  // Skip if binary already exists (e.g., CI caching)
  if (fs.existsSync(binPath)) {
    return;
  }

  console.log(`Downloading env-pilot v${VERSION}...`);

  const data = await download(url);
  fs.mkdirSync(binDir, { recursive: true });

  if (process.platform === "win32") {
    // Write zip and extract with PowerShell
    const zipPath = path.join(binDir, "env-pilot.zip");
    fs.writeFileSync(zipPath, data);
    execSync(
      `powershell -command "Expand-Archive -Force '${zipPath}' '${binDir}'"`,
      { stdio: "ignore" }
    );
    fs.unlinkSync(zipPath);
  } else {
    // Write tar.gz and extract
    const tarPath = path.join(binDir, "env-pilot.tar.gz");
    fs.writeFileSync(tarPath, data);
    execSync(`tar -xzf "${tarPath}" -C "${binDir}"`, { stdio: "ignore" });
    fs.unlinkSync(tarPath);
  }

  // Make binary executable
  if (process.platform !== "win32") {
    fs.chmodSync(binPath, 0o755);
  }

  console.log(`env-pilot v${VERSION} installed successfully.`);
}

main().catch((err) => {
  console.error(`Failed to install env-pilot: ${err.message}`);
  process.exit(1);
});
