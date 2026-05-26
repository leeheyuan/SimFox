import { contextBridge, ipcRenderer } from 'electron'

contextBridge.exposeInMainWorld('simfox', {
  pickInputFiles: () => ipcRenderer.invoke('local:pick-files'),
  pickFolder: () => ipcRenderer.invoke('local:pick-folder'),
  pickProjectFiles: () => ipcRenderer.invoke('local:pick-project-files'),
  readFileBytes: (targetPath: string) => ipcRenderer.invoke('local:read-file', targetPath),
  openPath: (targetPath: string) => ipcRenderer.invoke('local:open-path', targetPath),
  notify: (title: string, body: string) => ipcRenderer.invoke('local:notify', { title, body }),
  versions: {
    electron: process.versions.electron,
    chrome: process.versions.chrome,
  },
})
