import { readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";

const root = process.cwd();
const packageJson = JSON.parse(readFileSync(join(root, "package.json"), "utf8"));
const dependencies = { ...(packageJson.dependencies ?? {}), ...(packageJson.devDependencies ?? {}) };
const bannedPackages = ["axios", "zustand", "next-auth", "@auth/core"];
const foundPackages = bannedPackages.filter((name) => dependencies[name]);

if (foundPackages.length > 0) {
  throw new Error(`Banned packages detected: ${foundPackages.join(", ")}`);
}

const scanRoots = ["app", "features", "lib"];
const bannedPatterns = [/localStorage/g, /sessionStorage/g, /document\.cookie/g, /SessionProvider/g, /useSession/g];

function walk(directory) {
  for (const entry of readdirSync(directory)) {
    const fullPath = join(directory, entry);
    const stats = statSync(fullPath);

    if (stats.isDirectory()) {
      walk(fullPath);
      continue;
    }

    if (!/\.(ts|tsx|js|jsx)$/.test(entry)) {
      continue;
    }

    const source = readFileSync(fullPath, "utf8");
    for (const pattern of bannedPatterns) {
      if (pattern.test(source)) {
        throw new Error(`Banned pattern ${pattern} detected in ${fullPath}`);
      }
    }
  }
}

for (const relativeRoot of scanRoots) {
  walk(join(root, relativeRoot));
}

console.log("frontend guardrails verified");
