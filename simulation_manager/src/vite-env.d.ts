/// <reference types="vite/client" />
declare module "*.vue" {
  import type { DefineComponent } from "vue";
  const component: DefineComponent<{}, {}, any>;
  export default component;
}

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
  simfox?: {
    pickInputFiles: () => Promise<string[]>
    pickFolder: () => Promise<string | null>
    pickProjectFiles: () => Promise<SimfoxProjectBundle | null>
    readFileBytes: (targetPath: string) => Promise<number[]>
    openPath: (targetPath: string) => Promise<string>
    openInSumoGui: (configPath: string, projectName: string) => Promise<string>
    openInNetedit: (configPath: string, projectName: string) => Promise<string>
    notify: (title: string, body: string) => Promise<void>
    versions: {
      electron: string
      chrome: string
    }
  }
}
