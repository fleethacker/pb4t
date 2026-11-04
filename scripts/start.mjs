import { spawn, spawnSync } from "node:child_process";
import { existsSync, readFileSync, mkdirSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";

process.chdir(fileURLToPath(new URL("..", import.meta.url)));
if (existsSync(".env")) {
  for (const line of readFileSync(".env", "utf8").split(/\r?\n/)) {
    const match = line.match(/^\s*([A-Z_][A-Z0-9_]*)\s*=\s*(.*?)\s*$/);
    if (match && process.env[match[1]] === undefined)
      process.env[match[1]] = match[2].replace(/^(['"])(.*)\1$/, "$2");
  }
}
function run(command, args, cwd = process.cwd()) {
  const result = spawnSync(command, args, {
    cwd,
    stdio: "inherit",
    env: process.env,
  });
  if (result.error)
    console.error(
      `Could not run ${command}: ${result.error.message}. Install Go and Node.js 22.12+ first.`,
    );
  if (result.error || result.status !== 0) process.exit(result.status || 1);
}
mkdirSync("bin", { recursive: true });
if (!process.argv.includes("--reset")) {
  if (!existsSync("web/node_modules"))
    run("npm", ["install"], path.resolve("web"));
  run("npm", ["run", "build"], path.resolve("web"));
}
run("go", ["build", "-o", "bin/scheduler", "./cmd/scheduler"]);
const child = spawn(
  path.resolve("bin/scheduler"),
  process.argv.includes("--reset") ? ["--reset-demo"] : [],
  { stdio: "inherit", env: process.env },
);
for (const signal of ["SIGINT", "SIGTERM"])
  process.on(signal, () => child.kill(signal));
child.on("exit", (code) => process.exit(code ?? 0));
