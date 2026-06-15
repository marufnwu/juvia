import { useEffect, useState } from 'react'
import { Plus, Users as UsersIcon, Search, MoreHorizontal, Trash2, RefreshCw, Shield } from 'lucide-react'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { PageHeader } from '../../components/ui/Misc'
import { Modal, ConfirmModal } from '../../components/ui/Modal'
import { Input, Select } from '../../components/ui/Input'
import api from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'
import { cn } from '../../lib/utils'

interface User {
  id: number
  username: string
  email: string
  role: string
  active: boolean
  totp_enabled: boolean
  created_at: string
}

export default function Users() {
  const showError = useApiError()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [showMenu, setShowMenu] = useState<number | null>(null)

  const [showAddModal, setShowAddModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [showDeleteModal, setShowDeleteModal] = useState(false)

  const [addForm, setAddForm] = useState({ username: '', password: '', role: 'user', email: '' })
  const [editForm, setEditForm] = useState({ role: '', email: '', active: true })
  const [formLoading, setFormLoading] = useState(false)
  const [formError, setFormError] = useState('')

  useEffect(() => {
    loadUsers()
  }, [])

  const loadUsers = async () => {
    try {
      const res = await api.get('/users')
      setUsers(res.data.data || [])
    } catch (err: any) {
      showError(err, 'Failed to load users')
    } finally {
      setLoading(false)
    }
  }

  const filteredUsers = users.filter((u) =>
    u.username.toLowerCase().includes(search.toLowerCase()) ||
    u.email.toLowerCase().includes(search.toLowerCase())
  )

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormLoading(true)
    setFormError('')
    try {
      await api.post('/users', addForm)
      setShowAddModal(false)
      setAddForm({ username: '', password: '', role: 'user', email: '' })
      loadUsers()
    } catch (err: any) {
      setFormError(err.response?.data?.error?.user_message || 'Failed to create user')
    } finally {
      setFormLoading(false)
    }
  }

  const handleEdit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedUser) return
    setFormLoading(true)
    setFormError('')
    try {
      await api.put(`/users/${selectedUser.id}`, {
        role: editForm.role,
        email: editForm.email,
        active: editForm.active,
      })
      setShowEditModal(false)
      loadUsers()
    } catch (err: any) {
      setFormError(err.response?.data?.error?.user_message || 'Failed to update user')
    } finally {
      setFormLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!selectedUser) return
    setFormLoading(true)
    try {
      await api.delete(`/users/${selectedUser.id}`)
      setShowDeleteModal(false)
      setSelectedUser(null)
      loadUsers()
    } catch (err: any) {
      showError(err, 'Failed to delete user')
    } finally {
      setFormLoading(false)
    }
  }

  const openEdit = (user: User) => {
    setSelectedUser(user)
    setEditForm({ role: user.role, email: user.email, active: user.active })
    setShowEditModal(true)
    setShowMenu(null)
  }

  const openDelete = (user: User) => {
    setSelectedUser(user)
    setShowDeleteModal(true)
    setShowMenu(null)
  }

  const columns = [
    {
      key: 'username',
      header: 'Username',
      render: (u: User) => (
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 bg-primary/10 rounded flex items-center justify-center">
            <UsersIcon className="w-3.5 h-3.5 text-primary" />
          </div>
          <span className="font-medium text-foreground">{u.username}</span>
        </div>
      ),
    },
    { key: 'email', header: 'Email', render: (u: User) => <span className="text-text-secondary">{u.email}</span> },
    {
      key: 'role',
      header: 'Role',
      render: (u: User) => (
        <Badge variant={u.role === 'admin' ? 'pending' : 'neutral'}>
          {u.role === 'admin' ? (
            <><Shield className="w-3 h-3 mr-1" />Admin</>
          ) : (
            <><UsersIcon className="w-3 h-3 mr-1" />User</>
          )}
        </Badge>
      ),
    },
    {
      key: 'active',
      header: 'Status',
      render: (u: User) => (
        <Badge variant={u.active ? 'success' : 'danger'}>
          {u.active ? 'Active' : 'Inactive'}
        </Badge>
      ),
    },
    {
      key: 'totp_enabled',
      header: '2FA',
      render: (u: User) => (
        <Badge variant={u.totp_enabled ? 'success' : 'neutral'}>
          {u.totp_enabled ? 'Enabled' : 'Disabled'}
        </Badge>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      render: (u: User) => <span className="text-text-secondary">{formatDate(u.created_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      width: 'w-10',
      render: (u: User) => (
        <div className="relative">
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => { e.stopPropagation(); setShowMenu(showMenu === u.id ? null : u.id) }}
          >
            <MoreHorizontal className="w-4 h-4" />
          </Button>
          {showMenu === u.id && (
            <div className="absolute right-0 top-full mt-1 w-36 bg-surface border border-border rounded shadow-lg z-10">
              <button
                onClick={(e) => { e.stopPropagation(); openEdit(u) }}
                className="w-full flex items-center gap-2 px-3 py-2 text-sm text-foreground hover:bg-accent transition-colors"
              >
                Edit
              </button>
              <button
                onClick={(e) => { e.stopPropagation(); openDelete(u) }}
                className="w-full flex items-center gap-2 px-3 py-2 text-sm text-danger hover:bg-danger/10 transition-colors"
              >
                <Trash2 className="w-3.5 h-3.5" />
                Delete
              </button>
            </div>
          )}
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        title="User Management"
        description="Manage system users and their access levels"
        breadcrumbs={[{ label: 'Users' }]}
        actions={
          <Button onClick={() => setShowAddModal(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Add User
          </Button>
        }
      />

      <Card>
        <div className="flex items-center justify-between p-4 border-b border-border">
          <Input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search users..."
            className="w-64"
          />
          <Button variant="ghost" size="sm" onClick={loadUsers}>
            <RefreshCw className="w-4 h-4" />
          </Button>
        </div>

        <Table
          columns={columns}
          data={filteredUsers}
          keyField="id"
          loading={loading}
          emptyMessage={
            <div className="flex flex-col items-center gap-3 py-8">
              <UsersIcon className="w-10 h-10 text-text-secondary/50" />
              <p className="text-text-secondary text-sm">No users found</p>
              <Button size="sm" onClick={() => setShowAddModal(true)}>
                <Plus className="w-4 h-4 mr-2" />
                Add a user
              </Button>
            </div>
          }
        />
      </Card>

      <Modal
        open={showAddModal}
        onClose={() => setShowAddModal(false)}
        title="Add User"
        description="Create a new system user account"
      >
        <form onSubmit={handleAdd} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Username</label>
            <Input
              type="text"
              value={addForm.username}
              onChange={(e) => setAddForm({ ...addForm, username: e.target.value })}
              placeholder="e.g. johndoe"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Email</label>
            <Input
              type="email"
              value={addForm.email}
              onChange={(e) => setAddForm({ ...addForm, email: e.target.value })}
              placeholder="john@example.com"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Password</label>
            <Input
              type="password"
              value={addForm.password}
              onChange={(e) => setAddForm({ ...addForm, password: e.target.value })}
              placeholder="At least 8 characters"
              required
              minLength={8}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Role</label>
            <Select
              value={addForm.role}
              onChange={(e) => setAddForm({ ...addForm, role: e.target.value })}
            >
              <option value="user">User</option>
              <option value="admin">Admin</option>
            </Select>
          </div>
          {formError && (
            <div className="p-3 bg-danger/10 border border-danger/20 rounded text-sm text-danger">{formError}</div>
          )}
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setShowAddModal(false)}>Cancel</Button>
            <Button type="submit" loading={formLoading}>Create User</Button>
          </div>
        </form>
      </Modal>

      <Modal
        open={showEditModal}
        onClose={() => setShowEditModal(false)}
        title="Edit User"
        description={`Editing ${selectedUser?.username}`}
      >
        <form onSubmit={handleEdit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Email</label>
            <Input
              type="email"
              value={editForm.email}
              onChange={(e) => setEditForm({ ...editForm, email: e.target.value })}
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Role</label>
            <Select
              value={editForm.role}
              onChange={(e) => setEditForm({ ...editForm, role: e.target.value })}
            >
              <option value="user">User</option>
              <option value="admin">Admin</option>
            </Select>
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Status</label>
            <Select
              value={editForm.active ? 'true' : 'false'}
              onChange={(e) => setEditForm({ ...editForm, active: e.target.value === 'true' })}
            >
              <option value="true">Active</option>
              <option value="false">Inactive</option>
            </Select>
          </div>
          {formError && (
            <div className="p-3 bg-danger/10 border border-danger/20 rounded text-sm text-danger">{formError}</div>
          )}
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setShowEditModal(false)}>Cancel</Button>
            <Button type="submit" loading={formLoading}>Save Changes</Button>
          </div>
        </form>
      </Modal>

      <ConfirmModal
        open={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={handleDelete}
        title="Delete User"
        description={`Are you sure you want to delete "${selectedUser?.username}"? This action cannot be undone.`}
        confirmLabel="Delete"
        variant="danger"
        loading={formLoading}
      />
    </div>
  )
}
