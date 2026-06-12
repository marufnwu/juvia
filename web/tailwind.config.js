/** @type {import('tailwindcss').Config} */
export default {
  darkMode: "class",
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        // Design system colors
        primary: {
          DEFAULT: "#2563EB",
          50: "#EFF6FF",
          100: "#DBEAFE",
          200: "#BFDBFE",
          300: "#93C5FD",
          400: "#60A5FA",
          500: "#3B82F6",
          600: "#2563EB",
          700: "#1D4ED8",
          800: "#1E40AF",
          900: "#1E3A8A",
        },
        success: {
          DEFAULT: "#16A34A",
          50: "#F0FDF4",
          100: "#DCFCE7",
          200: "#BBF7D0",
          300: "#86EFAC",
          400: "#4ADE80",
          500: "#22C55E",
          600: "#16A34A",
          700: "#15803D",
          800: "#166534",
          900: "#14532D",
        },
        warning: {
          DEFAULT: "#D97706",
          50: "#FFFBEB",
          100: "#FEF3C7",
          200: "#FDE68A",
          300: "#FCD34D",
          400: "#FBBF24",
          500: "#F59E0B",
          600: "#D97706",
          700: "#B45309",
          800: "#92400E",
          900: "#78350F",
        },
        danger: {
          DEFAULT: "#DC2626",
          50: "#FEF2F2",
          100: "#FEE2E2",
          200: "#FECACA",
          300: "#FCA5A5",
          400: "#F87171",
          500: "#EF4444",
          600: "#DC2626",
          700: "#B91C1C",
          800: "#991B1B",
          900: "#7F1D1D",
        },
        neutral: "#6B7280",
        // Dark theme surfaces
        background: "#0F172A",
        surface: "#1E293B",
        border: "#334155",
        "text-primary": "#F8FAFC",
        "text-secondary": "#94A3B8",
        // shadcn/ui compatibility
        foreground: "#F8FAFC",
        muted: {
          DEFAULT: "#1E293B",
          foreground: "#94A3B8",
        },
        accent: {
          DEFAULT: "#334155",
          foreground: "#F8FAFC",
        },
        card: {
          DEFAULT: "#1E293B",
          foreground: "#F8FAFC",
        },
        popover: {
          DEFAULT: "#1E293B",
          foreground: "#F8FAFC",
        },
        destructive: {
          DEFAULT: "#DC2626",
          foreground: "#FFFFFF",
        },
        input: "#334155",
        ring: "#2563EB",
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "-apple-system", "sans-serif"],
        mono: ["JetBrains Mono", "Consolas", "monospace"],
      },
      borderRadius: {
        DEFAULT: "6px",
        card: "8px",
        tag: "4px",
      },
      boxShadow: {
        "glow-primary": "0 0 0 2px rgba(37, 99, 235, 0.3)",
        "glow-success": "0 0 0 2px rgba(22, 163, 74, 0.3)",
        "glow-danger": "0 0 0 2px rgba(220, 38, 38, 0.3)",
      },
      animation: {
        "pulse-slow": "pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite",
        "shimmer": "shimmer 2s infinite",
      },
      keyframes: {
        shimmer: {
          "0%": { backgroundPosition: "-200% 0" },
          "100%": { backgroundPosition: "200% 0" },
        },
      },
    },
  },
  plugins: [],
};