import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Database, Plus, Download, Table, Users, ChevronRight, Search, MoreHorizontal } from 'lucide-react'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { SearchInput } from '../../components/ui/SearchInput'
import { PageHeader } from '../../components/ui/Misc'
import { TableSkeleton } from '../../components/ui/Skeleton'
import api from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { cn } from '../../lib/utils'

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
    window.open(`/api/v1/databases/${selectedDB.id}/export`, '_blank')
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
          <button
            onClick={() => navigate('/databases/create')}
            className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white text-sm font-medium rounded hover:bg-primary/90 transition-colors"
          >
            <Plus className="w-4 h-4" />
            Create Database
          </button>
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
                No databases found
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
              <button
                onClick={exportDatabase}
                className="inline-flex items-center gap-2 px-3 py-2 border border-border text-sm rounded hover:bg-accent transition-colors"
              >
                <Download className="w-4 h-4" />
                Export
              </button>
            </div>

            <div className="flex border-b border-border">
              {[
                { id: 'tables', label: 'Tables', icon: Table },
                { id: 'users', label: 'Users', icon: Users },
                { id: 'sqleditor', label: 'SQL Editor', icon: Database },
              ].map((tab) => (
                <button
                  key={tab.id}
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
                </button>
              ))}
            </div>

            <div className="flex-1 overflow-auto p-4">
              {activeTab === 'tables' && (
                <div className="flex gap-4 h-full">
                  <div className="w-48 border border-border rounded overflow-auto">
                    {tables.map((t) => (
                      <button
                        key={t.name}
                        onClick={() => loadTableRows(t.name)}
                        className={cn(
                          'w-full text-left p-2.5 border-b border-border hover:bg-accent/50 transition-colors',
                          selectedTable === t.name && 'bg-accent'
                        )}
                      >
                        <div className="flex items-center gap-2">
                          <Table size={12} className="text-text-secondary" />
                          <span className="text-sm truncate">{t.name}</span>
                        </div>
                        <div className="text-xs text-text-secondary mt-0.5">
                          {t.rows.toLocaleString()} rows
                        </div>
                      </button>
                    ))}
                  </div>
                  <div className="flex-1 overflow-auto">
                    {selectedTable ? (
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="bg-accent/50">
                            {tableRows.length > 0 &&
                              Object.keys(tableRows[0]).map((col) => (
                                <th key={col} className="text-left p-2.5 border-b border-border font-medium text-text-secondary uppercase text-xs">
                                  {col}
                                </th>
                              ))}
                          </tr>
                        </thead>
                        <tbody>
                          {tableRows.map((row, i) => (
                            <tr key={i} className="border-b border-border hover:bg-accent/30">
                              {Object.values(row).map((val, j) => (
                                <td key={j} className="p-2.5 text-text-secondary font-mono text-xs">
                                  {String(val ?? 'NULL')}
                                </td>
                              ))}
                            </tr>
                          ))}
                        </tbody>
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
                    <h3 className="font-medium">Database Users</h3>
                    <button className="px-3 py-1.5 text-xs font-medium border border-border rounded hover:bg-accent transition-colors">
                      <Plus className="w-3.5 h-3.5 inline mr-1" />
                      Add User
                    </button>
                  </div>
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="bg-accent/50">
                        <th className="text-left p-2.5 border-b border-border font-medium text-text-secondary uppercase text-xs">Username</th>
                        <th className="text-left p-2.5 border-b border-border font-medium text-text-secondary uppercase text-xs">Host</th>
                        <th className="text-left p-2.5 border-b border-border font-medium text-text-secondary uppercase text-xs">Created</th>
                        <th className="text-left p-2.5 border-b border-border font-medium text-text-secondary uppercase text-xs">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {users.map((u) => (
                        <tr key={u.id} className="border-b border-border hover:bg-accent/30">
                          <td className="p-2.5 font-mono">{u.username}</td>
                          <td className="p-2.5 text-text-secondary">{u.host}</td>
                          <td className="p-2.5 text-text-secondary">{formatDate(u.created_at)}</td>
                          <td className="p-2.5">
                            <button className="text-xs text-danger hover:underline">Delete</button>
                          </td>
                        </tr>
                      ))}
                      {users.length === 0 && (
                        <tr>
                          <td colSpan={4} className="p-6 text-center text-text-secondary text-sm">
                            No users for this database
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
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
                    <button
                      onClick={runQuery}
                      className="px-4 py-2 bg-primary text-white text-sm font-medium rounded hover:bg-primary/90 transition-colors"
                    >
                      Run Query
                    </button>
                  </div>
                  {sqlError && (
                    <div className="p-3 bg-danger/10 border border-danger/20 rounded text-sm text-danger mb-3">
                      {sqlError}
                    </div>
                  )}
                  <div className="flex-1 overflow-auto border border-border rounded">
                    {sqlResults.length > 0 ? (
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="bg-accent/50">
                            {Object.keys(sqlResults[0]).map((col) => (
                              <th key={col} className="text-left p-2.5 border-b border-border font-medium text-text-secondary uppercase text-xs">
                                {col}
                              </th>
                            ))}
                          </tr>
                        </thead>
                        <tbody>
                          {sqlResults.map((row, i) => (
                            <tr key={i} className="border-b border-border hover:bg-accent/30">
                              {Object.values(row).map((val, j) => (
                                <td key={j} className="p-2.5 text-text-secondary font-mono text-xs">
                                  {String(val ?? 'NULL')}
                                </td>
                              ))}
                            </tr>
                          ))}
                        </tbody>
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
    </div>
  )
}