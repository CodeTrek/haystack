import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';

const haystackSupportedPlatforms = {
  "linux-x64": "linux-amd64",
  "linux-arm64": "linux-arm64",
  "darwin-x64": "darwin-amd64",
  "darwin-arm64": "darwin-arm64",
  "win32-x64": "windows-amd64",
  "win32-arm64": "windows-arm64",
}

const currentPlatform = `${process.platform}-${process.arch}`;
const isHaystackSupported = currentPlatform in haystackSupportedPlatforms;
type Status = 'initializing' | 'unsupported' | 'error' |'running';
type InstallStatus = 'checking' | 'installing' | 'downloading' | 'unsupported' | 'error' | 'installed' | 'not-installed';

export class Haystack {
  private corePath: string;
  private binDir: string;
  private status: Status;
  private installStatus: InstallStatus;
  constructor(private context: vscode.ExtensionContext) {
    // Use globalStorageUri for persistent storage across extension updates
    this.binDir = path.join(this.context.globalStorageUri.fsPath, 'bin');
    this.corePath = path.join(this.binDir, this.getExecutableName());
    this.status = 'initializing';
    this.installStatus = 'checking';
    this.start();
  }

  private getExecutableName(): string {
    // Adjust executable name based on the operating system
    return process.platform === 'win32' ? 'haystack.exe' : 'haystack';
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
   * Checks if the Haystack Core executable exists in the designated binary directory.
   * It first ensures the binary directory exists.
   * @returns True if the core executable exists, false otherwise.
   */
  public async start(): Promise<void> {
    if (!this.getIsSupported()) {
      this.status = 'unsupported';
      this.installStatus = 'unsupported';
      return;
    }

    try {
      // Ensure the target directory exists, creating it if necessary.
      await fs.promises.mkdir(this.binDir, { recursive: true });
      // Check for the existence of the core executable file.
      await fs.promises.access(this.corePath, fs.constants.F_OK);
      console.log(`Haystack Core found at: ${this.corePath}`);
      this.installStatus = 'installed';
    } catch (error) {
      // Log if the core executable is not found.
      console.log(`Haystack Core not found or accessible at: ${this.corePath}. Error: ${error}`);
      this.installStatus = 'not-installed';
    }

    if (this.installStatus === 'not-installed') {
      await this.install();
    }
  }

  /**
   * Retrieves the expected full path to the Haystack Core executable.
   */
  public getCorePath(): string {
    return this.corePath;
  }

    /**
   * Retrieves the directory where the Haystack Core binary should reside.
   */
  public getBinDir(): string {
    return this.binDir;
  }

  private async install(): Promise<void> {
    this.installStatus = 'installing';
    try {
    } catch (error) {
    }
  }
}
