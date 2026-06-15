import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Database, Plus, Download, Table as TableIcon, Users, ChevronRight, Search, MoreHorizontal, X } from 'lucide-react'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { SearchInput } from '../../components/ui/SearchInput'
import { PageHeader } from '../../components/ui/Misc'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '../../components/ui/Table'
import { TableSkeleton } from '../../components/ui/Skeleton'
import { Modal } from '../../components/ui/Modal'
import { Input } from '../../components/ui/Input'
import api from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { cn } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'

interface DatabaseInfo {
  id: number
  name: string
  engine: string
  user_id: number
  created_at: string
  size?: number
  table_count?: number
}

interface DBUser {
  id: number
  database_id: number
  username: string
  host: string
  created_at: string
}

interface TableInfo {
  name: string
  engine: string
  rows: number
  size: number
}

export default function DatabaseManager() {
  const navigate = useNavigate()
  const showError = useApiError()
  const [databases, setDatabases] = useState<DatabaseInfo[]>([])
  const [selectedDB, setSelectedDB] = useState<DatabaseInfo | null>(null)
  const [users, setUsers] = useState<DBUser[]>([])
  const [tables, setTables] = useState<TableInfo[]>([])
  const [activeTab, setActiveTab] = useState<'tables' | 'users' | 'sqleditor'>('tables')
  const [loading, setLoading] = useState(true)
  const [tableRows, setTableRows] = useState<Record<string, unknown>[]>([])
  const [selectedTable, setSelectedTable] = useState<string | null>(null)
  const [sqlQuery, setSqlQuery] = useState('SELECT * FROM ')
  const [sqlResults, setSqlResults] = useState<Record<string, unknown>[]>([])
  const [sqlError, setSqlError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [showAddUserModal, setShowAddUserModal] = useState(false)
  const [newUser, setNewUser] = useState({ username: '', password: '', host: 'localhost' })
  const [addingUser, setAddingUser] = useState(false)

  useEffect(() => {
    loadDatabases()
  }, [])

  const loadDatabases = async () => {
    setLoading(true)
    try {
      const res = await api.get('/databases')
      setDatabases(res.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const selectDatabase = async (db: DatabaseInfo) => {
    setSelectedDB(db)
    setActiveTab('tables')
    try {
      const [dbRes, tablesRes] = await Promise.allSettled([
        api.get(`/databases/${db.id}`),
        api.get(`/databases/${db.id}/tables`),
      ])

      if (dbRes.status === 'fulfilled') {
        setUsers(dbRes.value.data.data?.users || [])
      }
      if (tablesRes.status === 'fulfilled') {
        setTables(tablesRes.value.data.data?.tables || [])
      }
    } catch (err) {
      console.error(err)
    }
  }

  const loadTableRows = async (tableName: string) => {
    if (!selectedDB) return
    setSelectedTable(tableName)
    try {
      const res = await api.get(`/databases/${selectedDB.id}/tables/${tableName}/rows`)
      setTableRows(res.data.data?.rows || [])
    } catch (err) {
      console.error(err)
    }
  }

  const runQuery = async () => {
    if (!selectedDB) return
    setSqlError(null)
    try {
      const res = await api.post(`/databases/${selectedDB.id}/query`, { query: sqlQuery })
      setSqlResults(res.data.data?.rows || [])
    } catch (err: any) {
      setSqlError(err.response?.data?.message || 'Query failed')
      setSqlResults([])
    }
  }

  const exportDatabase = async () => {
    if (!selectedDB) return
    try {
      const res = await api.post(`/databases/${selectedDB.id}/export`, {}, {
        responseType: 'blob',
      })
      const blob = new Blob([res.data], { type: 'application/sql' })
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${selectedDB.name}.sql`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      window.URL.revokeObjectURL(url)
    } catch (err: any) {
      setSqlError(err.response?.data?.error?.user_message || 'Export failed')
    }
  }

  const handleAddUser = async () => {
    if (!selectedDB || !newUser.username || !newUser.password) return
    setAddingUser(true)
    try {
      await api.post(`/databases/${selectedDB.id}/users`, newUser)
      const res = await api.get(`/databases/${selectedDB.id}`)
      setUsers(res.data.data?.users || [])
      setShowAddUserModal(false)
      setNewUser({ username: '', password: '', host: 'localhost' })
    } catch (err: any) {
      showError(err, 'Failed to add user')
    } finally {
      setAddingUser(false)
    }
  }

  const handleDeleteUser = async (userId: number) => {
    if (!selectedDB) return
    try {
      await api.delete(`/databases/${selectedDB.id}/users/${userId}`)
      setUsers(users.filter((u) => u.id !== userId))
    } catch (err: any) {
      showError(err, 'Failed to delete user')
    }
  }

  const filteredDatabases = databases.filter((db) =>
    db.name.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="space-y-6">
      <PageHeader
        title="Databases"
        description="Manage MySQL and PostgreSQL databases"
        breadcrumbs={[{ label: 'Databases' }]}
        actions={
          <Button onClick={() => navigate('/databases/create')}>
            <Plus className="w-4 h-4 mr-2" />
            Create Database
          </Button>
        }
      />

      <div className="flex gap-6 h-[calc(100vh-14rem)]">
        <Card padding="none" className="w-72 flex flex-col">
          <div className="p-3 border-b border-border">
            <SearchInput
              value={search}
              onChange={setSearch}
              placeholder="Search databases..."
            />
          </div>
          <div className="flex-1 overflow-auto">
            {loading ? (
              <div className="p-4 space-y-2">
                {[...Array(5)].map((_, i) => (
                  <div key={i} className="h-14 skeleton rounded" />
                ))}
              </div>
            ) : filteredDatabases.length === 0 ? (
              <div className="p-4 text-center text-text-secondary text-sm">
                <Database className="w-10 h-10 mx-auto text-text-secondary/50 mb-3" />
                <p>No databases yet</p>
                <Button size="sm" className="mt-3" onClick={() => navigate('/databases/create')}>
                  Create your first database
                </Button>
              </div>
            ) : (
              filteredDatabases.map((db) => (
                <button
                  key={db.id}
                  onClick={() => selectDatabase(db)}
                  className={cn(
                    'w-full text-left p-3 border-b border-border hover:bg-accent/50 transition-colors',
                    selectedDB?.id === db.id && 'bg-accent'
                  )}
                >
                  <div className="flex items-center gap-2">
                    <Database size={14} className="text-text-secondary" />
                    <span className="text-sm font-medium truncate">{db.name}</span>
                  </div>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge variant="neutral">{db.engine}</Badge>
                    <span className="text-xs text-text-secondary">
                      {formatDate(db.created_at)}
                    </span>
                  </div>
                </button>
              ))
            )}
          </div>
        </Card>

        {selectedDB ? (
          <Card padding="none" className="flex-1 flex flex-col overflow-hidden">
            <div className="flex items-center justify-between p-4 border-b border-border">
              <div className="flex items-center gap-3">
                <Database className="w-5 h-5 text-primary" />
                <div>
                  <h2 className="font-semibold">{selectedDB.name}</h2>
                  <p className="text-xs text-text-secondary">{selectedDB.engine}</p>
                </div>
              </div>
              <Button variant="outline" size="sm" onClick={exportDatabase}>
                <Download className="w-4 h-4 mr-2" />
                Export
              </Button>
            </div>

            <div className="flex border-b border-border">
              {[
                { id: 'tables', label: 'Tables · Organized data rows', icon: TableIcon },
                { id: 'users', label: 'Users', icon: Users },
                { id: 'sqleditor', label: 'SQL Editor · Run database commands', icon: Database },
              ].map((tab) => (
                <Button
                  key={tab.id}
                  variant="ghost"
                  onClick={() => setActiveTab(tab.id as any)}
                  className={cn(
                    'flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors',
                    activeTab === tab.id
                      ? 'border-primary text-primary'
                      : 'border-transparent text-text-secondary hover:text-foreground'
                  )}
                >
                  <tab.icon className="w-4 h-4" />
                  {tab.label}
                </Button>
              ))}
            </div>

            <div className="flex-1 overflow-auto p-4">
              {activeTab === 'tables' && (
                <div className="flex gap-4 h-full">
                  <div className="w-48 border border-border rounded overflow-auto">
                    {tables.map((t) => (
                      <Button
                        key={t.name}
                        variant="ghost"
                        onClick={() => loadTableRows(t.name)}
                        className={cn(
                          'w-full justify-start p-2.5 border-b border-border hover:bg-accent/50 transition-colors h-auto text-left',
                          selectedTable === t.name && 'bg-accent'
                        )}
                      >
                        <div className="flex items-center gap-2">
                          <TableIcon size={12} className="text-text-secondary" />
                          <span className="text-sm truncate">{t.name}</span>
                        </div>
                        <div className="text-xs text-text-secondary mt-0.5">
                          {t.rows.toLocaleString()} rows
                        </div>
                      </Button>
                    ))}
                  </div>
                  <div className="flex-1 overflow-auto">
                    {selectedTable ? (
                      <table className="w-full">
                        <TableHeader>
                          <TableRow>
                            {tableRows.length > 0 &&
                              Object.keys(tableRows[0]).map((col) => (
                                <TableHead key={col}>{col}</TableHead>
                              ))}
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {tableRows.map((row, i) => (
                            <TableRow key={i}>
                              {Object.values(row).map((val, j) => (
                                <TableCell key={j} className="font-mono text-xs text-text-secondary">
                                  {String(val ?? 'NULL')}
                                </TableCell>
                              ))}
                            </TableRow>
                          ))}
                        </TableBody>
                      </table>
                    ) : (
                      <div className="text-center text-text-secondary py-8 text-sm">
                        Select a table to view rows
                      </div>
                    )}
                  </div>
                </div>
              )}

              {activeTab === 'users' && (
                <div>
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="font-medium">Database Users · Accounts for database access</h3>
                    <Button variant="outline" size="sm" onClick={() => setShowAddUserModal(true)}>
                      <Plus className="w-3.5 h-3.5 mr-1" />
                      Add User
                    </Button>
                  </div>
                  <Table
                    columns={[
                      { key: 'username', header: 'Username', render: (row: DBUser) => <span className="font-mono">{row.username}</span> },
                      { key: 'host', header: 'Host · Allowed computer (localhost = this server)', render: (row: DBUser) => <span className="text-text-secondary">{row.host}</span> },
                      { key: 'created_at', header: 'Created', render: (row: DBUser) => <span className="text-text-secondary">{formatDate(row.created_at)}</span> },
                      { key: 'actions', header: 'Actions', render: (row: DBUser) => (
                        <Button
                          variant="danger"
                          size="sm"
                          onClick={() => handleDeleteUser(row.id)}
                        >
                          Delete
                        </Button>
                      )},
                    ]}
                    data={users}
                    keyField="id"
                    emptyMessage={
                      <div className="flex flex-col items-center gap-3 py-8">
                        <Users className="w-10 h-10 text-text-secondary/50" />
                        <p className="text-text-secondary text-sm">No database users</p>
                        <Button size="sm" onClick={() => setShowAddUserModal(true)}>
                          Add a user
                        </Button>
                      </div>
                    }
                  />
                </div>
              )}

              {activeTab === 'sqleditor' && (
                <div className="h-full flex flex-col">
                  <textarea
                    value={sqlQuery}
                    onChange={(e) => setSqlQuery(e.target.value)}
                    className="w-full h-32 p-3 border border-border rounded bg-background text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary/50 resize-none"
                    placeholder="SELECT * FROM table_name LIMIT 100;"
                  />
                  <div className="flex items-center gap-2 my-3">
                    <Button onClick={runQuery}>
                      Run Query
                    </Button>
                  </div>
                  {sqlError && (
                    <div className="p-3 bg-danger/10 border border-danger/20 rounded text-sm text-danger mb-3">
                      {sqlError}
                    </div>
                  )}
                  <div className="flex-1 overflow-auto border border-border rounded">
                    {sqlResults.length > 0 ? (
                      <table className="w-full">
                        <TableHeader>
                          <TableRow>
                            {Object.keys(sqlResults[0]).map((col) => (
                              <TableHead key={col}>{col}</TableHead>
                            ))}
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {sqlResults.map((row, i) => (
                            <TableRow key={i}>
                              {Object.values(row).map((val, j) => (
                                <TableCell key={j} className="font-mono text-xs text-text-secondary">
                                  {String(val ?? 'NULL')}
                                </TableCell>
                              ))}
                            </TableRow>
                          ))}
                        </TableBody>
                      </table>
                    ) : (
                      <div className="p-6 text-center text-text-secondary text-sm">
                        Run a query to see results
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          </Card>
        ) : (
          <Card className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <Database className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
              <p className="text-text-secondary text-sm">Select a database to manage</p>
            </div>
          </Card>
        )}
      </div>

      <Modal open={showAddUserModal} onClose={() => setShowAddUserModal(false)} title="Add Database User">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Username</label>
            <Input
              value={newUser.username}
              onChange={(e) => setNewUser({ ...newUser, username: e.target.value })}
              placeholder="e.g. webapp_user"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Password</label>
            <Input
              type="password"
              value={newUser.password}
              onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
              placeholder="Minimum 8 characters"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Host · Which computer can connect</label>
            <Input
              value={newUser.host}
              onChange={(e) => setNewUser({ ...newUser, host: e.target.value })}
              placeholder="localhost"
            />
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowAddUserModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleAddUser} disabled={addingUser || !newUser.username || !newUser.password}>
              {addingUser ? 'Adding...' : 'Add User'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
