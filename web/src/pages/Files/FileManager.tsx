import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Folder, File, Upload, X, ChevronLeft, ChevronRight, Plus, Trash2,
  Download, Edit, Eye, Copy, RefreshCw, FolderPlus, FilePlus, MoreHorizontal,
  ArrowUp
} from 'lucide-react'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { Modal } from '../../components/ui/Modal'
import { Input, Label, FormGroup } from '../../components/ui/Input'
import { PageHeader } from '../../components/ui/Misc'
import api from '../../lib/api'
import { formatBytes, formatDate } from '../../lib/utils'
import { cn } from '../../lib/utils'

interface FileItem {
  name: string
  type: 'file' | 'directory'
  size: number
  modified_at: number
  permissions: string
}

export default function FileManager() {
  const { websiteId } = useParams<{ websiteId: string }>()
  const navigate = useNavigate()
  const [files, setFiles] = useState<FileItem[]>([])
  const [currentPath, setCurrentPath] = useState('')
  const [loading, setLoading] = useState(true)
  const [uploadProgress, setUploadProgress] = useState(false)
  const [editingFile, setEditingFile] = useState<string | null>(null)
  const [fileContent, setFileContent] = useState('')
  const [showHidden, setShowHidden] = useState(false)
  const [selectedFiles, setSelectedFiles] = useState<Set<string>>(new Set())
  const [showContextMenu, setShowContextMenu] = useState<{ x: number; y: number; file: FileItem } | null>(null)
  const [showNewFolderModal, setShowNewFolderModal] = useState(false)
  const [newFolderName, setNewFolderName] = useState('')

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
    setSelectedFiles(new Set())
  }

  const navigateUp = () => {
    if (!currentPath) return
    const parts = currentPath.split('/')
    parts.pop()
    const newPath = parts.join('/')
    loadFiles(newPath)
    setSelectedFiles(new Set())
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

  const toggleSelect = (fileName: string, e: React.MouseEvent) => {
    e.stopPropagation()
    const newSelected = new Set(selectedFiles)
    if (e.shiftKey && selectedFiles.size > 0) {
      // Range select
    } else if (newSelected.has(fileName)) {
      newSelected.delete(fileName)
    } else {
      newSelected.add(fileName)
    }
    setSelectedFiles(newSelected)
  }

  const handleContextMenu = (e: React.MouseEvent, file: FileItem) => {
    e.preventDefault()
    setShowContextMenu({ x: e.clientX, y: e.clientY, file })
  }

  const filteredFiles = showHidden ? files : files.filter((f) => !f.name.startsWith('.'))

  const fileIcon = (type: 'file' | 'directory') => {
    if (type === 'directory') return <Folder className="w-4 h-4 text-warning" />
    return <File className="w-4 h-4 text-text-secondary" />
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="File Manager"
        description={websiteId || ''}
        breadcrumbs={[
          { label: 'Files', href: '/files' },
          { label: websiteId || '' },
        ]}
      />

      <Card padding="none">
        <div className="p-3 flex items-center gap-3 border-b border-border">
          <button
            onClick={navigateUp}
            disabled={!currentPath}
            className="p-2 text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors disabled:opacity-50"
          >
            <ArrowUp className="w-4 h-4" />
          </button>

          <div className="flex items-center gap-2 text-sm">
            <button onClick={() => loadFiles('')} className="hover:text-primary transition-colors">
              root
            </button>
            {currentPath.split('/').map((part, i) => (
              <span key={i} className="flex items-center gap-2">
                <span className="text-text-secondary">/</span>
                <button
                  onClick={() => loadFiles(currentPath.split('/').slice(0, i + 1).join('/'))}
                  className="hover:text-primary transition-colors"
                >
                  {part}
                </button>
              </span>
            ))}
          </div>

          <div className="flex-1" />

          <label className="flex items-center gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={showHidden}
              onChange={(e) => setShowHidden(e.target.checked)}
              className="w-4 h-4 rounded border-border"
            />
            <span className="text-text-secondary">Show hidden</span>
          </label>

          <label className="flex items-center gap-2 px-3 py-2 bg-primary text-white text-sm rounded cursor-pointer hover:bg-primary/90 transition-colors">
            <Upload className="w-4 h-4" />
            Upload
            <input type="file" className="hidden" onChange={handleUpload} />
          </label>

          <button
            onClick={() => setShowNewFolderModal(true)}
            className="p-2 text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors"
          >
            <FolderPlus className="w-4 h-4" />
          </button>

          <button
            onClick={() => loadFiles(currentPath)}
            className="p-2 text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>

        {loading ? (
          <div className="p-8 text-center text-text-secondary">Loading...</div>
        ) : filteredFiles.length === 0 ? (
          <div className="p-8 text-center text-text-secondary">No files found</div>
        ) : (
          <div className="divide-y divide-border">
            {filteredFiles.map((file) => (
              <div
                key={file.name}
                onClick={() => file.type === 'directory' ? navigateToFolder(file.name) : handleEdit(file.name)}
                onContextMenu={(e) => handleContextMenu(e, file)}
                className={cn(
                  'flex items-center gap-3 p-3 hover:bg-accent/30 cursor-pointer transition-colors',
                  selectedFiles.has(file.name) && 'bg-primary/10'
                )}
              >
                <input
                  type="checkbox"
                  checked={selectedFiles.has(file.name)}
                  onChange={() => {}}
                  onClick={(e) => toggleSelect(file.name, e)}
                  className="w-4 h-4 rounded border-border"
                />
                <span className="w-5">{fileIcon(file.type)}</span>
                <span className="flex-1 text-sm font-medium truncate">{file.name}</span>
                <span className="text-xs text-text-secondary w-24 text-right">
                  {file.type === 'file' ? formatBytes(file.size) : '—'}
                </span>
                <span className="text-xs text-text-secondary w-40 text-right">
                  {formatDate(new Date(file.modified_at * 1000).toISOString())}
                </span>
                <div className="flex items-center gap-1">
                  {file.type === 'file' && (
                    <>
                      <button
                        onClick={(e) => { e.stopPropagation(); handleEdit(file.name) }}
                        className="p-1 text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors"
                      >
                        <Edit className="w-3.5 h-3.5" />
                      </button>
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDownload(file.name) }}
                        className="p-1 text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors"
                      >
                        <Download className="w-3.5 h-3.5" />
                      </button>
                    </>
                  )}
                  {(file.name.endsWith('.zip') || file.name.endsWith('.tar.gz')) && (
                    <button
                      onClick={(e) => { e.stopPropagation(); handleExtract(file.name) }}
                      className="p-1 text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors"
                    >
                      <Download className="w-3.5 h-3.5" />
                    </button>
                  )}
                  <button
                    onClick={(e) => { e.stopPropagation(); handleDelete(file.name) }}
                    className="p-1 text-text-secondary hover:text-danger hover:bg-danger/10 rounded transition-colors"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>

      {editingFile && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
          <div className="w-3/4 max-h-3/4 bg-surface border border-border rounded-card shadow-xl flex flex-col">
            <div className="flex items-center justify-between p-4 border-b border-border">
              <h3 className="font-semibold">{editingFile}</h3>
              <button onClick={() => setEditingFile(null)} className="p-1 text-text-secondary hover:text-foreground">
                <X className="w-4 h-4" />
              </button>
            </div>
            <textarea
              value={fileContent}
              onChange={(e) => setFileContent(e.target.value)}
              className="flex-1 p-4 bg-background font-mono text-sm resize-none focus:outline-none"
            />
            <div className="flex justify-end gap-3 p-4 border-t border-border">
              <Button variant="outline" onClick={() => setEditingFile(null)}>Cancel</Button>
              <Button onClick={handleSave}>Save</Button>
            </div>
          </div>
        </div>
      )}

      <Modal
        open={showNewFolderModal}
        onClose={() => setShowNewFolderModal(false)}
        title="Create Folder"
        size="sm"
      >
        <div className="space-y-4">
          <FormGroup>
            <Label>Folder Name</Label>
            <Input
              value={newFolderName}
              onChange={(e) => setNewFolderName(e.target.value)}
              placeholder="new_folder"
            />
          </FormGroup>
          <div className="flex justify-end gap-3">
            <Button variant="outline" onClick={() => setShowNewFolderModal(false)}>Cancel</Button>
            <Button onClick={() => setShowNewFolderModal(false)}>Create</Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}