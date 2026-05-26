declare module 'electron' {
  export const app: {
    whenReady(): Promise<void>
    on(event: string, listener: () => void): void
    quit(): void
  }

  export class BrowserWindow {
    constructor(options: unknown)
    loadURL(url: string): Promise<void>
    loadFile(filePath: string): Promise<void>
    static getAllWindows(): BrowserWindow[]
  }

  export const dialog: {
    showOpenDialog(options: unknown): Promise<{
      canceled: boolean
      filePaths: string[]
    }>
  }

  export const ipcMain: {
    handle(channel: string, listener: (...args: any[]) => unknown): void
  }

  export class Notification {
    constructor(options: { title: string; body: string })
    show(): void
  }

  export const shell: {
    openPath(targetPath: string): Promise<string>
  }

  export const contextBridge: {
    exposeInMainWorld(key: string, api: unknown): void
  }

  export const ipcRenderer: {
    invoke(channel: string, ...args: any[]): Promise<any>
  }
}
