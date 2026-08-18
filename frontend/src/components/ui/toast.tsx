"use client";

import { createContext, useContext, useState, useCallback, ReactNode } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { CheckCircle2, XCircle, AlertTriangle, Info, X } from "lucide-react";

type ToastType = "success" | "error" | "warning" | "info";

interface Toast {
  id: number;
  type: ToastType;
  title: string;
  message?: string;
}

interface ToastContextType {
  toast: (type: ToastType, title: string, message?: string) => void;
  success: (title: string, message?: string) => void;
  error: (title: string, message?: string) => void;
  warning: (title: string, message?: string) => void;
  info: (title: string, message?: string) => void;
}

const ToastContext = createContext<ToastContextType | null>(null);

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error("useToast must be used within ToastProvider");
  return ctx;
}

const iconMap = {
  success: CheckCircle2,
  error: XCircle,
  warning: AlertTriangle,
  info: Info,
};

const styleMap = {
  success: {
    bg: "bg-emerald-50 dark:bg-emerald-950/50",
    border: "border-emerald-200 dark:border-emerald-800",
    icon: "text-emerald-600",
    title: "text-emerald-800 dark:text-emerald-200",
    message: "text-emerald-600 dark:text-emerald-400",
    progress: "bg-emerald-500",
  },
  error: {
    bg: "bg-red-50 dark:bg-red-950/50",
    border: "border-red-200 dark:border-red-800",
    icon: "text-red-600",
    title: "text-red-800 dark:text-red-200",
    message: "text-red-600 dark:text-red-400",
    progress: "bg-red-500",
  },
  warning: {
    bg: "bg-amber-50 dark:bg-amber-950/50",
    border: "border-amber-200 dark:border-amber-800",
    icon: "text-amber-600",
    title: "text-amber-800 dark:text-amber-200",
    message: "text-amber-600 dark:text-amber-400",
    progress: "bg-amber-500",
  },
  info: {
    bg: "bg-blue-50 dark:bg-blue-950/50",
    border: "border-blue-200 dark:border-blue-800",
    icon: "text-blue-600",
    title: "text-blue-800 dark:text-blue-200",
    message: "text-blue-600 dark:text-blue-400",
    progress: "bg-blue-500",
  },
};

let toastId = 0;

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const removeToast = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const addToast = useCallback(
    (type: ToastType, title: string, message?: string) => {
      const id = ++toastId;
      setToasts((prev) => [...prev, { id, type, title, message }]);
      setTimeout(() => removeToast(id), 4000);
    },
    [removeToast]
  );

  const contextValue: ToastContextType = {
    toast: addToast,
    success: (title, message) => addToast("success", title, message),
    error: (title, message) => addToast("error", title, message),
    warning: (title, message) => addToast("warning", title, message),
    info: (title, message) => addToast("info", title, message),
  };

  return (
    <ToastContext.Provider value={contextValue}>
      {children}

      {/* Toast Container */}
      <div className="fixed top-4 right-4 z-[100] flex flex-col gap-3 w-full max-w-sm pointer-events-none">
        <AnimatePresence mode="popLayout">
          {toasts.map((t) => {
            const Icon = iconMap[t.type];
            const styles = styleMap[t.type];

            return (
              <motion.div
                key={t.id}
                layout
                initial={{ opacity: 0, x: 80, scale: 0.95 }}
                animate={{ opacity: 1, x: 0, scale: 1 }}
                exit={{ opacity: 0, x: 80, scale: 0.95 }}
                transition={{ type: "spring", duration: 0.4, bounce: 0.15 }}
                className={`pointer-events-auto relative overflow-hidden rounded-xl border ${styles.bg} ${styles.border} shadow-lg backdrop-blur-sm`}
              >
                <div className="flex items-start gap-3 p-4">
                  <div className="shrink-0 mt-0.5">
                    <Icon className={`h-5 w-5 ${styles.icon}`} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className={`text-sm font-semibold ${styles.title}`}>{t.title}</p>
                    {t.message && (
                      <p className={`text-xs mt-0.5 ${styles.message} line-clamp-2`}>{t.message}</p>
                    )}
                  </div>
                  <button
                    onClick={() => removeToast(t.id)}
                    className="shrink-0 text-muted-foreground/50 hover:text-muted-foreground transition-colors"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
                {/* Progress bar */}
                <motion.div
                  initial={{ width: "100%" }}
                  animate={{ width: "0%" }}
                  transition={{ duration: 4, ease: "linear" }}
                  className={`h-0.5 ${styles.progress} opacity-40`}
                />
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>
    </ToastContext.Provider>
  );
}
