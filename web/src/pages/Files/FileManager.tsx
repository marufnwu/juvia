import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Folder, File, Upload, X, ChevronLeft, ChevronRight, Plus, Trash2,
  Download, Edit, Eye, Copy, RefreshCw, FolderPlus, FilePlus, MoreHorizontal,
  ArrowUp, Pencil, FolderOpen
} from 'lucide-react'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { Modal, ConfirmModal } from '../../components/ui/Modal'
import { Input, Label, FormGroup, Switch } from '../../components/ui/Input'
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
  const [showRenameModal, setShowRenameModal] = useState(false)
  const [renamingFile, setRenamingFile] = useState<FileItem | null>(null)
  const [newFileName, setNewFileName] = useState('')
  const [creatingFolder, setCreatingFolder] = useState(false)
  const [renaming, setRenaming] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)
  const [deleting, setDeleting] = useState(false)

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
    const file = e.target.files[0]
    const formData = new FormData()
    formData.append('path', currentPath)
    formData.append('file_name', file.name)
    formData.append('content', file)
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

  const handleDelete = (fileName: string) => {
    setConfirmDelete(fileName)
  }

  const doDelete = async () => {
    if (!confirmDelete || !websiteId) return
    const fileName = confirmDelete
    setConfirmDelete(null)
    setDeleting(true)
    try {
      await api.delete(`/websites/${websiteId}/files/delete`, {
        data: { path: currentPath ? `${currentPath}/${fileName}` : fileName },
      })
      loadFiles(currentPath)
    } catch (err) {
      console.error(err)
    } finally {
      setDeleting(false)
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
      const rawContent = data.content || ''
      try {
        setFileContent(atob(rawContent))
      } catch {
        setFileContent(rawContent)
      }
      setEditingFile(fileName)
    } catch (err) {
      console.error(err)
    }
  }

  const handleSave = async () => {
    if (!websiteId || !editingFile) return
    const path = currentPath ? `${currentPath}/${editingFile}` : editingFile
    try {
      await api.put(`/websites/${websiteId}/files/edit`, { path, content: btoa(fileContent) })
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

  const handleCreateFolder = async () => {
    if (!websiteId || !newFolderName) return
    setCreatingFolder(true)
    try {
      const path = currentPath ? `${currentPath}/${newFolderName}` : newFolderName
      await api.post(`/websites/${websiteId}/files/folder`, { path })
      setShowNewFolderModal(false)
      setNewFolderName('')
      loadFiles(currentPath)
    } catch (err) {
      console.error(err)
    } finally {
      setCreatingFolder(false)
    }
  }

  const openRenameModal = (file: FileItem) => {
    setRenamingFile(file)
    setNewFileName(file.name)
    setShowRenameModal(true)
  }

  const handleRename = async () => {
    if (!websiteId || !renamingFile || !newFileName) return
    setRenaming(true)
    try {
      const oldPath = currentPath ? `${currentPath}/${renamingFile.name}` : renamingFile.name
      const newPath = currentPath ? `${currentPath}/${newFileName}` : newFileName
      await api.put(`/websites/${websiteId}/files/rename`, { old_path: oldPath, new_path: newPath })
      setShowRenameModal(false)
      setRenamingFile(null)
      setNewFileName('')
      loadFiles(currentPath)
    } catch (err) {
      console.error(err)
    } finally {
      setRenaming(false)
    }
  }

  const toggleSelect = (fileName: string, e: React.MouseEvent) => {
    e.stopPropagation()
    const newSelected = new Set(selectedFiles)
    if (e.shiftKey && selectedFiles.size > 0) {
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
          <Button variant="ghost" size="sm" onClick={navigateUp} disabled={!currentPath}>
            <ArrowUp className="w-4 h-4" />
          </Button>

          <div className="flex items-center gap-2 text-sm">
            <Button variant="ghost" size="sm" onClick={() => loadFiles('')}>
              root
            </Button>
            {currentPath.split('/').map((part, i) => (
              <span key={i} className="flex items-center gap-2">
                <span className="text-text-secondary">/</span>
                <Button variant="ghost" size="sm" onClick={() => loadFiles(currentPath.split('/').slice(0, i + 1).join('/'))}>
                  {part}
                </Button>
              </span>
            ))}
          </div>

          <div className="flex-1" />

          <Switch
            checked={showHidden}
            onChange={(e) => setShowHidden(e.target.checked)}
            label="Show hidden"
          />

          <label className="flex items-center gap-2 px-3 py-2 bg-primary text-white text-sm rounded cursor-pointer hover:bg-primary/90 transition-colors">
            <Upload className="w-4 h-4" />
            Upload
            <input type="file" className="hidden" onChange={handleUpload} />
          </label>

          <Button variant="outline" size="sm" onClick={() => setShowNewFolderModal(true)}>
            <FolderPlus className="w-4 h-4" />
          </Button>

          <Button variant="outline" size="sm" onClick={() => loadFiles(currentPath)}>
            <RefreshCw className="w-4 h-4" />
          </Button>
        </div>

        {loading ? (
          <div className="p-8 text-center text-text-secondary">Loading...</div>
        ) : filteredFiles.length === 0 ? (
          <div className="p-8 text-center text-text-secondary">
            <FolderOpen className="w-10 h-10 mx-auto text-text-secondary/50 mb-3" />
            <p>This directory is empty</p>
          </div>
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
                      <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleEdit(file.name) }}>
                        <Edit className="w-3.5 h-3.5" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleDownload(file.name) }}>
                        <Download className="w-3.5 h-3.5" />
                      </Button>
                    </>
                  )}
                  <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); openRenameModal(file) }} title="Rename">
                    <Pencil className="w-3.5 h-3.5" />
                  </Button>
                  {(file.name.endsWith('.zip') || file.name.endsWith('.tar.gz')) && (
                    <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleExtract(file.name) }}>
                      <Download className="w-3.5 h-3.5" />
                    </Button>
                  )}
                  <Button variant="danger" size="sm" onClick={(e) => { e.stopPropagation(); handleDelete(file.name) }}>
                    <Trash2 className="w-3.5 h-3.5" />
                  </Button>
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
              <Button variant="ghost" onClick={() => setEditingFile(null)}>
                <X className="w-4 h-4" />
              </Button>
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
            <Button onClick={handleCreateFolder} loading={creatingFolder}>Create</Button>
          </div>
        </div>
      </Modal>

      <Modal open={showRenameModal} onClose={() => setShowRenameModal(false)} title="Rename File" size="sm">
        <div className="space-y-4">
          {renamingFile && (
            <>
              <FormGroup>
                <Label>Current Name</Label>
                <Input value={renamingFile.name} disabled />
              </FormGroup>
              <FormGroup>
                <Label>New Name</Label>
                <Input
                  value={newFileName}
                  onChange={(e) => setNewFileName(e.target.value)}
                  placeholder="new_name"
                />
              </FormGroup>
              <div className="flex justify-end gap-3">
                <Button variant="outline" onClick={() => setShowRenameModal(false)}>Cancel</Button>
                <Button onClick={handleRename} loading={renaming} disabled={!newFileName || newFileName === renamingFile.name}>Rename</Button>
              </div>
            </>
          )}
        </div>
      </Modal>

      <ConfirmModal
        open={!!confirmDelete}
        onClose={() => setConfirmDelete(null)}
        onConfirm={doDelete}
        title="Delete File"
        description={confirmDelete ? `Delete ${confirmDelete}?` : ''}
        confirmLabel="Delete"
        variant="danger"
        loading={deleting}
      />
    </div>
  )
}
