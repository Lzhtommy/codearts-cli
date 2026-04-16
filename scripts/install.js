// codearts-cli postinstall — downloads the prebuilt Go binary from GitHub
// Releases for the current platform/arch.
//
// Flow: npm install → postinstall hook → this script
//   1. Detect OS + arch (darwin/linux/windows × amd64/arm64)
//   2. Download tarball/zip from GitHub Releases (with mirror fallback)
//   3. Extract binary to ./bin/codearts-cli[.exe]
//
// If behind a firewall, set https_proxy before running npm install.

const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const os = require("os");

const VERSION = require("../package.json").version;
const REPO = "Lzhtommy/codearts-cli";
const NAME = "codearts-cli";

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

const platform = PLATFORM_MAP[process.platform];
const arch = ARCH_MAP[process.arch];

if (!platform || !arch) {
  console.error(
    `Unsupported platform: ${process.platform}-${process.arch}`
  );
  process.exit(1);
}

const isWindows = process.platform === "win32";
const ext = isWindows ? ".zip" : ".tar.gz";
const archiveName = `${NAME}-${VERSION}-${platform}-${arch}${ext}`;
const GITHUB_URL = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;
const MIRROR_URL = `https://registry.npmmirror.com/-/binary/${NAME}/v${VERSION}/${archiveName}`;

const binDir = path.join(__dirname, "..", "bin");
const dest = path.join(binDir, NAME + (isWindows ? ".exe" : ""));

fs.mkdirSync(binDir, { recursive: true });

// If the binary already exists (e.g. placed by `make npm-link` or a local
// build), skip the download entirely. This avoids a network round-trip and
// makes `npm link` work offline during development.
if (fs.existsSync(dest)) {
  try {
    const { execSync: exec } = require("child_process");
    exec(`"${dest}" --version`, { stdio: "ignore" });
    console.log(`${NAME} binary already present at ${dest} — skipping download`);
    process.exit(0);
  } catch (_) {
    // Binary exists but is broken (wrong arch, corrupt, etc.) — re-download.
  }
}

function download(url, destPath) {
  const sslFlag = isWindows ? "--ssl-revoke-best-effort " : "";
  execSync(
    `curl ${sslFlag}--fail --location --silent --show-error --connect-timeout 10 --max-time 120 --output "${destPath}" "${url}"`,
    { stdio: ["ignore", "ignore", "pipe"] }
  );
}

function install() {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), `${NAME}-`));
  const archivePath = path.join(tmpDir, archiveName);

  try {
    // Try GitHub first, fall back to npmmirror for China-mainland users.
    try {
      console.log(`Downloading ${NAME} v${VERSION} (${platform}-${arch})...`);
      download(GITHUB_URL, archivePath);
    } catch (_) {
      console.log("GitHub download failed, trying mirror...");
      download(MIRROR_URL, archivePath);
    }

    if (isWindows) {
      execSync(
        `powershell -Command "Expand-Archive -Path '${archivePath}' -DestinationPath '${tmpDir}'"`,
        { stdio: "ignore" }
      );
    } else {
      execSync(`tar -xzf "${archivePath}" -C "${tmpDir}"`, {
        stdio: "ignore",
      });
    }

    const binaryName = NAME + (isWindows ? ".exe" : "");
    const extractedBinary = path.join(tmpDir, binaryName);

    fs.copyFileSync(extractedBinary, dest);
    fs.chmodSync(dest, 0o755);
    console.log(`${NAME} v${VERSION} installed successfully`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

try {
  install();
} catch (err) {
  console.error(`Failed to install ${NAME}:`, err.message);
  console.error(
    `\nIf you are behind a firewall or in a restricted network, try:\n` +
    `  export https_proxy=http://your-proxy:port\n` +
    `  npm install -g @Lzhtommy/codearts-cli\n\n` +
    `Or install from source:\n` +
    `  git clone https://github.com/${REPO}.git && cd ${NAME}\n` +
    `  make install PREFIX=$HOME/.local`
  );
  process.exit(1);
}
