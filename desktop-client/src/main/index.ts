import { app, BrowserWindow, dialog, ipcMain, Notification, shell } from 'electron'
import { dirname, join, basename, extname, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { promises as fs } from 'node:fs'

const currentDir = dirname(fileURLToPath(import.meta.url))
const devServerUrl = process.env.SIMFOX_DESKTOP_URL
const rendererDistPath = resolve(currentDir, '../../../simulation_manager/dist/index.html')

function createWindow() {
  const mainWindow = new BrowserWindow({
    title: 'SimFox Desktop',
    width: 1280,
    height: 860,
    minWidth: 960,
    minHeight: 640,
    webPreferences: {
      preload: join(currentDir, '../preload/index.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: false,
    },
  })

  if (devServerUrl) {
    void mainWindow.loadURL(devServerUrl)
  } else {
    void mainWindow.loadFile(rendererDistPath)
  }
}

ipcMain.handle('local:pick-files', async () => {
  const result = await dialog.showOpenDialog({
    properties: ['openFile', 'multiSelections'],
    filters: [
      { name: 'SUMO inputs', extensions: ['sumocfg', 'xml', 'json', 'csv'] },
      { name: 'All files', extensions: ['*'] },
    ],
  })

  return result.canceled ? [] : result.filePaths
})

ipcMain.handle('local:pick-folder', async () => {
  const result = await dialog.showOpenDialog({
    properties: ['openDirectory'],
  })

  return result.canceled ? null : result.filePaths[0]
})

ipcMain.handle('local:pick-project-files', async () => {
  const result = await dialog.showOpenDialog({
    properties: ['openFile'],
    filters: [{ name: 'Project Files', extensions: ['sumocfg', 'xml'] }],
  })

  if (result.canceled || result.filePaths.length === 0) {
    return null
  }

  const selectedPath = result.filePaths[0]
  if (!isSupportedImportFile(selectedPath)) {
    throw new Error('Only .sumocfg or .net.xml files can be imported')
  }

  return buildProjectBundle(selectedPath)
})

ipcMain.handle('local:read-file', async (_event: unknown, targetPath: string) => {
  const bytes = await fs.readFile(targetPath)
  return Array.from(bytes)
})

ipcMain.handle('local:open-path', async (_event: unknown, targetPath: string) => {
  return shell.openPath(targetPath)
})

ipcMain.handle('local:notify', (_event: unknown, payload: { title: string; body: string }) => {
  new Notification({ title: payload.title, body: payload.body }).show()
})

app.whenReady().then(createWindow)

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow()
  }
})

async function buildProjectBundle(selectedPath: string) {
  const absoluteSelectedPath = resolve(selectedPath)
  const filesByAbsolutePath = new Map<string, { absolutePath: string; relativePath: string; name: string }>()
  const missing: string[] = []

  filesByAbsolutePath.set(absoluteSelectedPath, {
    absolutePath: absoluteSelectedPath,
    relativePath: basename(absoluteSelectedPath),
    name: basename(absoluteSelectedPath),
  })

  if (absoluteSelectedPath.toLowerCase().endsWith('.sumocfg')) {
    const configDirectory = dirname(absoluteSelectedPath)
    const configContents = await fs.readFile(absoluteSelectedPath, 'utf8')
    const dependencyPaths = extractDependencies(configContents)

    for (const dependencyPath of dependencyPaths) {
      const absoluteDependencyPath = resolve(configDirectory, dependencyPath)
      try {
        await fs.access(absoluteDependencyPath)
        filesByAbsolutePath.set(absoluteDependencyPath, {
          absolutePath: absoluteDependencyPath,
          relativePath: normalizePath(relative(configDirectory, absoluteDependencyPath)),
          name: basename(absoluteDependencyPath),
        })
      } catch {
        missing.push(dependencyPath)
      }
    }
  }

  return {
    projectName: basename(absoluteSelectedPath, extname(absoluteSelectedPath)),
    configPath: absoluteSelectedPath.toLowerCase().endsWith('.sumocfg') ? absoluteSelectedPath : '',
    baseDirectory: dirname(absoluteSelectedPath),
    files: Array.from(filesByAbsolutePath.values()),
    missing,
  }
}

function isSupportedImportFile(value: string) {
  const normalized = value.toLowerCase()
  return normalized.endsWith('.sumocfg') || normalized.endsWith('.net.xml')
}

function extractDependencies(configText: string): string[] {
  const keys = ['net-file', 'route-files', 'additional-files', 'gui-settings-file']
  const values = keys.flatMap((key) => {
    const pattern = new RegExp(`<${key}[^>]*value="([^"]+)"[^>]*/?>`, 'g')
    const matches: string[] = []
    let match: RegExpExecArray | null
    while ((match = pattern.exec(configText)) !== null) {
      matches.push(match[1])
    }
    return matches
  })

  return values
    .flatMap((value) => value.split(','))
    .map((value) => normalizePath(value))
    .filter((value, index, array) => value !== '' && array.indexOf(value) === index)
}

function normalizePath(value: string) {
  return value.replace(/\\/g, '/').replace(/^\.?\//, '').trim()
}
