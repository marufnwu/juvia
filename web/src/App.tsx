import { BrowserRouter, Routes, Route } from 'react-router-dom'
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
import Settings from './pages/Settings/Settings'
import Terminal from './pages/Terminal/Terminal'
import Metrics from './pages/Metrics/Metrics'
import FilesIndex from './pages/Files/FilesIndex'
import Webmail from './pages/Webmail/Webmail'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="websites" element={<WebsiteList />} />
          <Route path="websites/create" element={<CreateWebsite />} />
          <Route path="websites/:id" element={<WebsiteDetail />} />
          <Route path="websites/:id/logs" element={<LogViewer />} />
          <Route path="websites/trash" element={<Trash />} />
          <Route path="databases" element={<DatabaseManager />} />
          <Route path="databases/create" element={<CreateDatabase />} />
          <Route path="email" element={<EmailDashboard />} />
          <Route path="email/webmail" element={<Webmail />} />
          <Route path="files" element={<FilesIndex />} />
          <Route path="files/:websiteId" element={<FileManager />} />
          <Route path="firewall" element={<Firewall />} />
          <Route path="backups" element={<Backups />} />
          <Route path="cron" element={<CronJobs />} />
          <Route path="alerts" element={<Alerts />} />
          <Route path="metrics" element={<Metrics />} />
          <Route path="terminal" element={<Terminal />} />
          <Route path="terminal/recordings" element={<Recordings />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App