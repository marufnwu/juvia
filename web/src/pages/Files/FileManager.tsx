import { useEffect, useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { Folder, File, Upload, X } from 'lucide-react'
import api from '../../lib/api'

interface FileItem {
  name: string
  type: 'file' | 'directory'
  size: number
  modified_at: number
  permissions: string
}

export default function FileManager() {
  const { websiteId } = useParams<{ websiteId: string }>()
  const [files, setFiles] = useState<FileItem[]>([])
  const [currentPath, setCurrentPath] = useState('')
  const [loading, setLoading] = useState(true)
  const [uploadProgress, setUploadProgress] = useState(false)
  const [editingFile, setEditingFile] = useState<string | null>(null)
  const [fileContent, setFileContent] = useState('')
  const [showHidden, setShowHidden] = useState(false)
  const [selectedFiles, setSelectedFiles] = useState<Set<string>>(new Set())

  const loadFiles = useCallback(async (path: string) => {
    if (!websiteId) return
    setLoading(true)
    try {
      const res = await api.get(`/websites/${websiteId}/files?path=${encodeURIComponent(path)}`)
      const data = res.data.data as { items: FileItem[] }
      setFiles(data.items || [])
      setCurrentPath(path)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [websiteId])

  useEffect(() => {
    loadFiles('')
  }, [loadFiles])

  const navigateToFolder = (folderName: string) => {
    const newPath = currentPath ? `${currentPath}/${folderName}` : folderName
    loadFiles(newPath)
  }

  const navigateUp = () => {
    if (!currentPath) return
    const parts = currentPath.split('/')
    parts.pop()
    const newPath = parts.join('/')
    loadFiles(newPath)
  }

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files?.length || !websiteId) return
    setUploadProgress(true)
    const formData = new FormData()
    formData.append('path', currentPath)
    formData.append('file_name', e.target.files[0].name)
    formData.append('content', await e.target.files[0].text())
    try {
      await api.post(`/websites/${websiteId}/files/upload`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      loadFiles(currentPath)
    } catch (err) {
      console.error(err)
    } finally {
      setUploadProgress(false)
    }
  }

  const handleDelete = async (fileName: string) => {
    if (!confirm(`Delete ${fileName}?`)) return
    try {
      await api.delete(`/websites/${websiteId}/files/delete`, {
        data: { path: currentPath ? `${currentPath}/${fileName}` : fileName },
      })
      loadFiles(currentPath)
    } catch (err) {
      console.error(err)
    }
  }

  const handleDownload = async (fileName: string) => {
    if (!websiteId) return
    const path = currentPath ? `${currentPath}/${fileName}` : fileName
    window.open(`/api/v1/websites/${websiteId}/files/download?path=${encodeURIComponent(path)}`, '_blank')
  }

  const handleEdit = async (fileName: string) => {
    if (!websiteId) return
    const path = currentPath ? `${currentPath}/${fileName}` : fileName
    try {
      const res = await api.get(`/websites/${websiteId}/files/edit?path=${encodeURIComponent(path)}`)
      const data = res.data.data as { content: string }
      setFileContent(data.content || '')
      setEditingFile(fileName)
    } catch (err) {
      console.error(err)
    }
  }

  const handleSave = async () => {
    if (!websiteId || !editingFile) return
    const path = currentPath ? `${currentPath}/${editingFile}` : editingFile
    try {
      await api.put(`/websites/${websiteId}/files/edit`, { path, content: fileContent })
      setEditingFile(null)
    } catch (err) {
      console.error(err)
    }
  }

  const handleExtract = async (fileName: string) => {
    if (!websiteId) return
    const path = currentPath ? `${currentPath}/${fileName}` : fileName
    try {
      await api.post(`/websites/${websiteId}/files/extract`, { path })
      loadFiles(currentPath)
    } catch (err) {
      console.error(err)
    }
  }

  const toggleSelect = (fileName: string) => {
    const newSelected = new Set(selectedFiles)
    if (newSelected.has(fileName)) {
      newSelected.delete(fileName)
    } else {
      newSelected.add(fileName)
    }
    setSelectedFiles(newSelected)
  }

  const filteredFiles = showHidden
    ? files
    : files.filter((f) => !f.name.startsWith('.'))

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <h1 className="text-lg font-semibold">File Manager</h1>
          {currentPath && (
            <div className="flex items-center gap-2">
              <button onClick={navigateUp} className="text-sm text-muted-foreground hover:text-foreground">
                ← Up
              </button>
              <span className="text-sm text-muted-foreground">/ {currentPath}</span>
            </div>
          )}
        </div>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={showHidden}
              onChange={(e) => setShowHidden(e.target.checked)}
            />
            Show hidden
          </label>
        </div>
      </div>

      <div className="flex items-center gap-4 mb-6">
        <label className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 text-sm font-medium cursor-pointer hover:opacity-90">
          <Upload size={16} />
          Upload
          <input type="file" className="hidden" onChange={handleUpload} />
        </label>
        {uploadProgress && <span className="text-sm text-muted-foreground">Uploading...</span>}
      </div>

      {loading ? (
        <div className="text-muted-foreground">Loading files...</div>
      ) : filteredFiles.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <p>No files found</p>
        </div>
      ) : (
        <div className="border border-border">
          <table className="w-full text-sm">
            <thead className="bg-muted/50 border-b border-border">
              <tr>
                <th className="text-left p-3 w-8"></th>
                <th className="text-left p-3 font-medium">Name</th>
                <th className="text-left p-3 font-medium">Type</th>
                <th className="text-left p-3 font-medium">Size</th>
                <th className="text-left p-3 font-medium">Modified</th>
                <th className="text-left p-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredFiles.map((file) => (
                <tr key={file.name} className="border-b border-border hover:bg-muted/30">
                  <td className="p-3">
                    <input
                      type="checkbox"
                      checked={selectedFiles.has(file.name)}
                      onChange={() => toggleSelect(file.name)}
                    />
                  </td>
                  <td className="p-3">
                    {file.type === 'directory' ? (
                      <button
                        onClick={() => navigateToFolder(file.name)}
                        className="flex items-center gap-2 text-primary hover:underline"
                      >
                        <Folder size={16} className="text-yellow-500" />
                        {file.name}
                      </button>
                    ) : (
                      <span className="flex items-center gap-2">
                        <File size={16} className="text-muted-foreground" />
                        {file.name}
                      </span>
                    )}
                  </td>
                  <td className="p-3 text-muted-foreground capitalize">
                    {file.type === 'directory' ? 'Folder' : 'File'}
                  </td>
                  <td className="p-3 text-muted-foreground">
                    {file.type === 'file' ? formatSize(file.size) : '—'}
                  </td>
                  <td className="p-3 text-muted-foreground">
                    {new Date(file.modified_at * 1000).toLocaleString()}
                  </td>
                  <td className="p-3">
                    <div className="flex gap-2">
                      {file.type === 'file' && (
                        <>
                          <button onClick={() => handleEdit(file.name)} className="text-primary hover:underline text-xs">
                            Edit
                          </button>
                          <button onClick={() => handleDownload(file.name)} className="text-primary hover:underline text-xs">
                            Download
                          </button>
                        </>
                      )}
                      {(file.name.endsWith('.zip') || file.name.endsWith('.tar.gz')) && (
                        <button onClick={() => handleExtract(file.name)} className="text-primary hover:underline text-xs">
                          Extract
                        </button>
                      )}
                      <button onClick={() => handleDelete(file.name)} className="text-red-600 hover:underline text-xs">
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {editingFile && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border border-border p-6 w-3/4 max-h-3/4 overflow-auto">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-medium">{editingFile}</h3>
              <button onClick={() => setEditingFile(null)} className="text-muted-foreground hover:text-foreground">
                <X size={20} />
              </button>
            </div>
            <textarea
              value={fileContent}
              onChange={(e) => setFileContent(e.target.value)}
              className="w-full h-96 border border-border p-2 font-mono text-sm"
            />
            <div className="flex gap-2 mt-4">
              <button
                onClick={handleSave}
                className="bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
              >
                Save
              </button>
              <button
                onClick={() => setEditingFile(null)}
                className="border border-border px-4 py-2 text-sm hover:bg-muted"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}