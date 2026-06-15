import { z } from 'zod'

export const domainSchema = z.string().min(1, 'Domain is required').refine(
  (val) => /^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/.test(val),
  'Invalid domain name'
)

export const emailSchema = z.string().min(1, 'Email is required').email('Invalid email address')

export const cronSchema = z.string().min(1, 'Cron expression is required').refine(
  (val) => /^[0-9*,/-]+\s+[0-9*,/-]+\s+[0-9*,/-]+\s+[0-9*,/-]+\s+[0-9*,/-]+$/.test(val.trim()),
  'Invalid cron expression'
)

export const portSchema = z.string().min(1, 'Port is required').refine(
  (val) => /^[0-9]+$/.test(val) && parseInt(val) >= 1 && parseInt(val) <= 65535,
  'Port must be between 1 and 65535'
)

export const cidrSchema = z.string().min(1, 'CIDR is required').refine(
  (val) => /^([0-9]{1,3}\.){3}[0-9]{1,3}\/[0-9]{1,2}$/.test(val),
  'Invalid CIDR notation (e.g., 192.168.1.0/24)'
)

export const usernameSchema = z.string().min(1, 'Username is required').min(3, 'Username must be at least 3 characters').max(32, 'Username must be at most 32 characters').regex(
  /^[a-zA-Z0-9_-]+$/,
  'Username can only contain letters, numbers, underscores, and hyphens'
)

export const passwordSchema = z.string().min(1, 'Password is required').min(8, 'Password must be at least 8 characters')

export const hostnameSchema = z.string().min(1, 'Hostname is required').refine(
  (val) => /^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$/,
  'Invalid hostname'
)

export const pathSchema = z.string().min(1, 'Path is required').refine(
  (val) => val.startsWith('/'),
  'Path must start with /'
)

export const urlSchema = z.string().min(1, 'URL is required').url('Invalid URL')

export const gitRepoSchema = z.string().min(1, 'Repository URL is required').refine(
  (val) => /^https?:\/\/.+\.git$/.test(val) || /^https?:\/\/github\.com\/.+\/.+$/.test(val),
  'Invalid git repository URL'
)

export type Domain = z.infer<typeof domainSchema>
export type Email = z.infer<typeof emailSchema>
export type Cron = z.infer<typeof cronSchema>
export type Port = z.infer<typeof portSchema>
export type CIDR = z.infer<typeof cidrSchema>
export type Username = z.infer<typeof usernameSchema>
export type Password = z.infer<typeof passwordSchema>
export type Hostname = z.infer<typeof hostnameSchema>
export type Path = z.infer<typeof pathSchema>
export type GitRepo = z.infer<typeof gitRepoSchema>
