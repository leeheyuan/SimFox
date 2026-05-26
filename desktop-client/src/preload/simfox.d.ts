export {}

declare global {
  interface SimfoxProjectFile {
    absolutePath: string
    relativePath: string
    name: string
  }

  interface SimfoxProjectBundle {
    projectName: string
    configPath: string
    baseDirectory: string
    files: SimfoxProjectFile[]
    missing: string[]
  }

  interface Window {
    simfox: {
      pickInputFiles: () => Promise<string[]>
      pickFolder: () => Promise<string | null>
      pickProjectFiles: () => Promise<SimfoxProjectBundle | null>
      readFileBytes: (targetPath: string) => Promise<number[]>
      openPath: (targetPath: string) => Promise<string>
      notify: (title: string, body: string) => Promise<void>
      versions: {
        electron: string
        chrome: string
      }
    }
  }
}
