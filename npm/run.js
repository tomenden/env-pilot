#!/usr/bin/env node
"use strict";

const { execFileSync } = require("child_process");
const path = require("path");

const binName = process.platform === "win32" ? "env-pilot.exe" : "env-pilot";
const binPath = path.join(__dirname, "bin", binName);

try {
  execFileSync(binPath, process.argv.slice(2), { stdio: "inherit" });
} catch (err) {
  if (err.status !== undefined) {
    process.exit(err.status);
  }
  console.error(`Failed to run env-pilot: ${err.message}`);
  process.exit(1);
}
