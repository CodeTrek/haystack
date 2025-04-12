import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';

const supportedPlatforms = {
  "linux-amd64": true,
  "linux-arm64": true,
  "darwin-amd64": true,
  "darwin-arm64": true,
  "windows-amd64": true,
  "windows-arm64": true,
}

const HAYSTACK_LOCAL_INSTALL_PORT= 13135;

const haystackConfig= (context: vscode.ExtensionContext) => `
global:
  data_path: ${path.join(context.globalStorageUri.fsPath, 'data')}
  port: ${HAYSTACK_LOCAL_INSTALL_PORT}
`

const platform = () => {
  if (process.platform === 'win32') {
    return 'windows';
  }
  return process.platform;
}

const arch = () => {
  if (process.arch === 'x64') {
    return 'amd64';
  }
  return process.arch;
}

const currentPlatform = `${platform()}-${arch()}`;
const isHaystackSupported = supportedPlatforms[currentPlatform as keyof typeof supportedPlatforms] || false;
const HAYSTACK_DOWNLOAD_URL = 'https://github.com/CodeTrek/haystack/releases/download/';
const HAYSTACK_DOWNLOAD_URL_FALLBACK = 'https://haystack.codetrek.cn/download/';
const HAYSTACK_VERSION = 'v1.0.0';
const HAYSTACK_ZIP_FILE_NAME = `haystack-${currentPlatform}-${HAYSTACK_VERSION}.zip`;

type Status = 'initializing' | 'unsupported' | 'error' |'running';
type InstallStatus = 'checking' | 'downloading' | 'unsupported' | 'error' | 'installed' | 'not-installed';

export class Haystack {
  private coreFilePath: string;
  private binDir: string;
  private status: Status;
  private installStatus: InstallStatus;
  private downloadZipPath: string;
  private builtinZipPath: string;

  constructor(private context: vscode.ExtensionContext) {
    // Use globalStorageUri for persistent storage across extension updates
    this.binDir = this.context.globalStorageUri.fsPath;
    this.coreFilePath = path.join(this.binDir, this.getExecutableName());
    this.downloadZipPath = path.join(this.context.globalStorageUri.fsPath, "download");
    this.builtinZipPath = path.join(this.context.extensionPath, "pkgs"); // We may have builtin zip files
    this.status = 'initializing';
    this.installStatus = 'checking';
    this.doInit();
  }

  public getIsSupported(): boolean {
    return isHaystackSupported;
  }

  public getCurrentPlatform(): string {
    return currentPlatform;
  }

  public getStatus(): Status {
    return this.status;
  }

  public getInstallStatus(): InstallStatus {
    return this.installStatus;
  }

  /**
   * Retrieves the expected full path to the Haystack Core executable.
   */
  public getCorePath(): string {
    return this.coreFilePath;
  }

    /**
   * Retrieves the directory where the Haystack Core binary should reside.
   */
  public getBinDir(): string {
    return this.binDir;
  }

  private getExecutableName(): string {
    return platform() === 'windows' ? 'haystack.exe' : 'haystack';
  }

  /**
   * Checks if the Haystack Core executable exists in the designated binary directory.
   * It first ensures the binary directory exists.
   * @returns True if the core executable exists, false otherwise.
   */
  private async doInit(): Promise<void> {
    if (!this.getIsSupported()) {
      this.status = 'unsupported';
      this.installStatus = 'unsupported';
      return;
    }

    try {
      // Ensure the target directory exists, creating it if necessary.
      await fs.promises.mkdir(this.binDir, { recursive: true });
      // Check for the existence of the core executable file.
      await fs.promises.access(this.coreFilePath, fs.constants.F_OK);
      console.log(`Haystack Core found at: ${this.coreFilePath}`);
      this.installStatus = 'installed';
    } catch (error) {
      // Log if the core executable is not found.
      console.log(`Haystack Core not found or accessible at: ${this.coreFilePath}.`);
      this.installStatus = 'not-installed';
    }

    if (this.installStatus === 'not-installed') {
      await this.install();
    }

    if (this.installStatus !== 'installed') {
      this.status = 'error';
      return
    }

    await this.start();
  }

  private async install(): Promise<void> {
    try {
      await fs.promises.mkdir(this.downloadZipPath, { recursive: true });

      const downloadedZipPath = path.join(this.downloadZipPath, HAYSTACK_ZIP_FILE_NAME);
      const builtinZipFilePath = path.join(this.builtinZipPath, HAYSTACK_ZIP_FILE_NAME);

      // Step 1: Check if zip already exists in download directory
      try {
        await this.checkExistingZip(downloadedZipPath, true);
        console.log('Found downloaded zip file');
        await this.extractZip(downloadedZipPath);
        this.installStatus = 'installed';
        return;
      } catch (error) {
        console.log('No downloaded zip file found');
      }

      // Step 2: Check if zip exists in builtin directory
      try {
        await this.checkExistingZip(builtinZipFilePath, false);
        console.log('Found builtin zip file');
        await this.extractZip(builtinZipFilePath);
        this.installStatus = 'installed';
        return;
      } catch (error) {
        console.log('No builtin zip file found');
      }

      // Step 3: Try downloading from primary URL
      this.installStatus = 'downloading';
      try {
        const primaryUrl = `${HAYSTACK_DOWNLOAD_URL}${HAYSTACK_VERSION}/${HAYSTACK_ZIP_FILE_NAME}`;
        console.log(`Downloading from primary URL: ${primaryUrl}`);
        await this.downloadFile(primaryUrl, downloadedZipPath);
        await this.extractZip(downloadedZipPath);
        this.installStatus = 'installed';
        return;
      } catch (error) {
        console.log(`Failed to download from primary URL: ${error}`);
      }

      // Step 4: Try downloading from fallback URL
      try {
        const fallbackUrl = `${HAYSTACK_DOWNLOAD_URL_FALLBACK}${HAYSTACK_VERSION}/${HAYSTACK_ZIP_FILE_NAME}`;
        console.log(`Downloading from fallback URL: ${fallbackUrl}`);
        await this.downloadFile(fallbackUrl, downloadedZipPath);
        await this.extractZip(downloadedZipPath);
        this.installStatus = 'installed';
        return;
      } catch (error) {
        console.log(`Failed to download from fallback URL: ${error}`);
        this.installStatus = 'error';
      }
    } catch (error) {
      console.error(`Installation failed: ${error}`);
      this.installStatus = 'error';
    }
  }

  private async downloadFile(url: string, destination: string): Promise<void> {
    const https = require('https');
    const http = require('http');

    return new Promise((resolve, reject) => {
      const client = url.startsWith('https') ? https : http;
      const file = fs.createWriteStream(destination);

      client.get(url, (response: any) => {
        if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
          file.close();
          // Handle redirects
          fs.unlink(destination, () => {
            this.downloadFile(response.headers.location, destination)
              .then(resolve)
              .catch(reject);
          });
          return;
        }

        if (response.statusCode !== 200) {
          file.close();
          fs.unlink(destination, () => {});
          reject(new Error(`Failed to download, status code: ${response.statusCode}`));
          return;
        }

        response.pipe(file);

        file.on('finish', () => {
          file.close();
          resolve();
        });
      }).on('error', (err: Error) => {
        file.close();
        fs.unlink(destination, () => {});
        reject(err);
      });

      file.on('error', (err: Error) => {
        file.close();
        fs.unlink(destination, () => {});
        reject(err);
      });
    });
  }

  private async checkExistingZip(zipFilePath: string, deleteIfSmall: boolean = false): Promise<void> {
    try {
      // Check if file exists
      await fs.promises.access(zipFilePath, fs.constants.F_OK);

      // Check file size
      const stats = await fs.promises.stat(zipFilePath);
      const minSizeBytes = 1024 * 1024; // 1MB in bytes

      if (stats.size < minSizeBytes) {
        if (deleteIfSmall) {
          await fs.promises.unlink(zipFilePath);
          console.log(`Deleted small zip file: ${zipFilePath}`);
        }

        throw new Error(`Zip file exists but is too small (${stats.size} bytes, minimum: ${minSizeBytes} bytes)`);
      }

      console.log(`Verified zip file: ${zipFilePath}`);
    } catch (error) {
      if (error instanceof Error) {
        throw error; // Re-throw existing errors
      } else {
        throw new Error(`Zip file does not exist or cannot be accessed: ${zipFilePath}`);
      }
    }
  }

  private async extractZip(zipFilePath: string): Promise<void> {
    // Use a zip extraction method compatible with your environment
    const AdmZip = require('adm-zip');
    const zip = new AdmZip(zipFilePath);

    return new Promise((resolve, reject) => {
      try {
        zip.extractAllTo(this.binDir, true);

        // Make the executable file runnable on non-Windows platforms
        if (platform() !== 'windows') {
          fs.chmodSync(this.coreFilePath, 0o755);
        }

        resolve();
      } catch (error) {
        reject(error);
      }
    });
  }

  private async start(): Promise<void> {
    this.status = 'running';
  }
}
