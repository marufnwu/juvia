import { BrowserRouter, Routes, Route, Navigate, useParams } from 'react-router-dom'
import { QueryProvider } from './lib/queryClient'
import Layout from './components/layout/Layout'
import Dashboard from './pages/Dashboard/Dashboard'
import Login from './pages/Auth/Login'
import WebsiteList from './pages/Websites/WebsiteList'
import CreateWebsite from './pages/Websites/CreateWebsite'
import WebsiteDetail from './pages/Websites/WebsiteDetail'
import Trash from './pages/Websites/Trash'
import DatabaseManager from './pages/Databases/DatabaseManager'
import CreateDatabase from './pages/Databases/CreateDatabase'
import EmailDashboard from './pages/Email/EmailDashboard'
import FileManager from './pages/Files/FileManager'
import Firewall from './pages/Firewall/Firewall'
import Backups from './pages/Backups/Backups'
import CronJobs from './pages/Cron/CronJobs'
import Alerts from './pages/Alerts/Alerts'
import LogViewer from './pages/Logs/LogViewer'
import Recordings from './pages/Terminal/Recordings'
import AuditLog from './pages/AuditLog/AuditLog'
import Settings from './pages/Settings/Settings'
import Terminal from './pages/Terminal/Terminal'
import Metrics from './pages/Metrics/Metrics'
import FilesIndex from './pages/Files/FilesIndex'
import Webmail from './pages/Webmail/Webmail'
import Setup from './pages/Setup/Setup'
import Users from './pages/Users/Users'
import DnsManagement from './pages/DNS/DnsManagement'
import NotFound from './pages/Errors/NotFound'
import { ToastContainer } from './components/ui/Toast'
import { ErrorBoundary } from './components/ErrorBoundary'
import { useAuthStore } from './stores/authStore'
import { useEffect } from 'react'

function WebsiteAccessLogs() {
  const { id } = useParams()
  return <LogViewer websiteId={Number(id)} logType="access" />
}

function WebsiteErrorLogs() {
  const { id } = useParams()
  return <LogViewer websiteId={Number(id)} logType="error" />
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

function AppRoutes() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const refreshUser = useAuthStore((s) => s.refreshUser)

  useEffect(() => {
    if (isAuthenticated) {
      refreshUser()
    }
  }, [isAuthenticated, refreshUser])

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/setup" element={<Setup />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="websites" element={<WebsiteList />} />
        <Route path="websites/create" element={<CreateWebsite />} />
        <Route path="websites/:id" element={<WebsiteDetail />} />
        <Route path="websites/:id/logs/access" element={<WebsiteAccessLogs />} />
        <Route path="websites/:id/logs/error" element={<WebsiteErrorLogs />} />
        <Route path="websites/trash" element={<Trash />} />
        <Route path="databases" element={<DatabaseManager />} />
        <Route path="databases/create" element={<CreateDatabase />} />
        <Route path="email" element={<EmailDashboard />} />
        <Route path="email/webmail" element={<Webmail />} />
        <Route path="files" element={<FilesIndex />} />
        <Route path="files/:websiteId" element={<FileManager />} />
        <Route path="firewall" element={<Firewall />} />
        <Route path="dns" element={<DnsManagement />} />
        <Route path="backups" element={<Backups />} />
        <Route path="cron" element={<CronJobs />} />
        <Route path="alerts" element={<Alerts />} />
        <Route path="metrics" element={<Metrics />} />
        <Route path="terminal" element={<Terminal />} />
        <Route path="terminal/recordings" element={<Recordings />} />
        <Route path="audit-log" element={<AuditLog />} />
        <Route path="settings" element={<Settings />} />
        <Route path="users" element={<Users />} />
      </Route>
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <QueryProvider>
          <AppRoutes />
          <ToastContainer />
        </QueryProvider>
      </BrowserRouter>
    </ErrorBoundary>
  )
}

export default App