import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Database, Plus, Table, Download } from 'lucide-react'
import api from '../../lib/api'

interface Database {
  id: number
  name: string
  engine: string
  user_id: number
  created_at: string
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
}

export default function DatabaseManager() {
  const navigate = useNavigate()
  const [databases, setDatabases] = useState<Database[]>([])
  const [selectedDB, setSelectedDB] = useState<Database | null>(null)
  const [users, setUsers] = useState<DBUser[]>([])
  const [tables, setTables] = useState<TableInfo[]>([])
  const [activeTab, setActiveTab] = useState<'tables' | 'users' | 'sqleditor'>('tables')
  const [loading, setLoading] = useState(true)
  const [tableRows, setTableRows] = useState<Record<string, unknown>[]>([])
  const [selectedTable, setSelectedTable] = useState<string | null>(null)
  const [sqlQuery, setSqlQuery] = useState('SELECT * FROM ')
  const [sqlResults, setSqlResults] = useState<Record<string, unknown>[]>([])
  const [sqlError, setSqlError] = useState<string | null>(null)

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

  const selectDatabase = async (db: Database) => {
    setSelectedDB(db)
    setActiveTab('tables')
    try {
      const res = await api.get(`/databases/${db.id}`)
      setUsers(res.data.data.users || [])
      const tablesRes = await api.get(`/databases/${db.id}/tables`)
      setTables(tablesRes.data.data.tables || [])
    } catch (err) {
      console.error(err)
    }
  }

  const loadTableRows = async (tableName: string) => {
    if (!selectedDB) return
    setSelectedTable(tableName)
    try {
      const res = await api.get(`/databases/${selectedDB.id}/tables/${tableName}/rows`)
      setTableRows(res.data.data.rows || [])
    } catch (err) {
      console.error(err)
    }
  }

  const runQuery = async () => {
    if (!selectedDB) return
    setSqlError(null)
    try {
      const res = await api.post(`/databases/${selectedDB.id}/query`, { query: sqlQuery })
      setSqlResults(res.data.data.rows || [])
    } catch (err: any) {
      setSqlError(err.response?.data?.message || 'Query failed')
      setSqlResults([])
    }
  }

  const exportDatabase = async () => {
    if (!selectedDB) return
    window.open(`/api/v1/databases/${selectedDB.id}/export`, '_blank')
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-lg font-semibold">Databases</h1>
        <button
          onClick={() => navigate('/databases/create')}
          className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
        >
          <Plus size={16} />
          Create Database
        </button>
      </div>

      <div className="flex gap-6 h-[calc(100vh-12rem)]">
        <div className="w-72 border border-border overflow-auto">
          <div className="p-3 font-medium text-sm border-b border-border">MySQL & PostgreSQL</div>
          {loading ? (
            <div className="p-4 text-sm text-muted-foreground">Loading...</div>
          ) : databases.length === 0 ? (
            <div className="p-4 text-sm text-muted-foreground">No databases</div>
          ) : (
            databases.map((db) => (
              <button
                key={db.id}
                onClick={() => selectDatabase(db)}
                className={`w-full text-left p-3 border-b border-border hover:bg-muted/50 ${
                  selectedDB?.id === db.id ? 'bg-muted' : ''
                }`}
              >
                <div className="flex items-center gap-2">
                  <Database size={14} />
                  <span className="text-sm font-medium truncate">{db.name}</span>
                </div>
                <div className="flex items-center gap-2 mt-1">
                  <span className="text-xs text-muted-foreground">{db.engine}</span>
                </div>
              </button>
            ))
          )}
        </div>

        {selectedDB ? (
          <div className="flex-1 border border-border overflow-hidden flex flex-col">
            <div className="flex border-b border-border">
              <button
                onClick={() => setActiveTab('tables')}
                className={`px-4 py-2 text-sm font-medium ${activeTab === 'tables' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
              >
                Tables
              </button>
              <button
                onClick={() => setActiveTab('users')}
                className={`px-4 py-2 text-sm font-medium ${activeTab === 'users' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
              >
                Users
              </button>
              <button
                onClick={() => setActiveTab('sqleditor')}
                className={`px-4 py-2 text-sm font-medium ${activeTab === 'sqleditor' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
              >
                SQL Editor
              </button>
              <div className="flex-1" />
              <button
                onClick={exportDatabase}
                className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground flex items-center gap-1"
              >
                <Download size={14} />
                Export
              </button>
            </div>

            <div className="flex-1 overflow-auto p-4">
              {activeTab === 'tables' && (
                <div className="flex gap-4 h-full">
                  <div className="w-48 border border-border overflow-auto">
                    {tables.map((t) => (
                      <button
                        key={t.name}
                        onClick={() => loadTableRows(t.name)}
                        className={`w-full text-left p-2 border-b border-border hover:bg-muted/50 ${
                          selectedTable === t.name ? 'bg-muted' : ''
                        }`}
                      >
                        <div className="flex items-center gap-2">
                          <Table size={12} />
                          <span className="text-sm truncate">{t.name}</span>
                        </div>
                      </button>
                    ))}
                  </div>
                  <div className="flex-1 overflow-auto">
                    {selectedTable ? (
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="bg-muted/50">
                            {tableRows.length > 0 &&
                              Object.keys(tableRows[0]).map((col) => (
                                <th key={col} className="text-left p-2 border-b border-border font-medium">
                                  {col}
                                </th>
                              ))}
                          </tr>
                        </thead>
                        <tbody>
                          {tableRows.map((row, i) => (
                            <tr key={i} className="border-b border-border">
                              {Object.values(row).map((val, j) => (
                                <td key={j} className="p-2 text-muted-foreground">
                                  {String(val)}
                                </td>
                              ))}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    ) : (
                      <div className="text-sm text-muted-foreground">Select a table</div>
                    )}
                  </div>
                </div>
              )}

              {activeTab === 'users' && (
                <div>
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="bg-muted/50">
                        <th className="text-left p-2 border-b border-border font-medium">Username</th>
                        <th className="text-left p-2 border-b border-border font-medium">Host</th>
                        <th className="text-left p-2 border-b border-border font-medium">Created</th>
                      </tr>
                    </thead>
                    <tbody>
                      {users.map((u) => (
                        <tr key={u.id} className="border-b border-border">
                          <td className="p-2">{u.username}</td>
                          <td className="p-2 text-muted-foreground">{u.host}</td>
                          <td className="p-2 text-muted-foreground">{new Date(u.created_at).toLocaleDateString()}</td>
                        </tr>
                      ))}
                      {users.length === 0 && (
                        <tr>
                          <td colSpan={3} className="p-4 text-center text-muted-foreground">No users</td>
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
                    className="w-full h-32 border border-border p-2 font-mono text-sm mb-2"
                    placeholder="Enter SQL query..."
                  />
                  <button
                    onClick={runQuery}
                    className="bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90 mb-4 w-fit"
                  >
                    Run Query
                  </button>
                  {sqlError && (
                    <div className="text-red-600 text-sm mb-2">{sqlError}</div>
                  )}
                  <div className="flex-1 overflow-auto border border-border">
                    {sqlResults.length > 0 ? (
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="bg-muted/50">
                            {Object.keys(sqlResults[0]).map((col) => (
                              <th key={col} className="text-left p-2 border-b border-border font-medium">
                                {col}
                              </th>
                            ))}
                          </tr>
                        </thead>
                        <tbody>
                          {sqlResults.map((row, i) => (
                            <tr key={i} className="border-b border-border">
                              {Object.values(row).map((val, j) => (
                                <td key={j} className="p-2 text-muted-foreground">
                                  {String(val)}
                                </td>
                              ))}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    ) : (
                      <div className="p-4 text-sm text-muted-foreground">Results will appear here</div>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="flex-1 border border-border flex items-center justify-center text-muted-foreground">
            Select a database
          </div>
        )}
      </div>
    </div>
  )
}