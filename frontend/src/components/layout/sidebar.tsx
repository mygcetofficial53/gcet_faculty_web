"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  Calendar,
  ClipboardCheck,
  Zap,
  BarChart3,
  User,
  Settings,
  HelpCircle,
  Heart,
  Info,
  LogOut,
  GraduationCap
} from "lucide-react";
import { useAuthStore } from "@/store/useAuthStore";
import { api } from "@/lib/api";
import { useRouter } from "next/navigation";

const routes = [
  { label: "Dashboard", icon: LayoutDashboard, href: "/" },
  { label: "Timetable", icon: Calendar, href: "/timetable" },
  { label: "Attendance", icon: ClipboardCheck, href: "/attendance" },
  { label: "Analytics", icon: BarChart3, href: "/analytics" },
];

const bottomRoutes = [
  { label: "Profile", icon: User, href: "/profile" },
  { label: "Settings", icon: Settings, href: "/settings" },
  { label: "Support", icon: HelpCircle, href: "/support" },
  { label: "Support Dev", icon: Heart, href: "/support-developer" },
  { label: "About", icon: Info, href: "/about" },
];

interface SidebarProps {
  onNavigate?: () => void;
}

export function Sidebar({ onNavigate }: SidebarProps) {
  const pathname = usePathname();
  const router = useRouter();
  const clearAuth = useAuthStore((state) => state.clearAuth);

  const handleLogout = async () => {
    try {
      await api.post("/auth/logout");
    } finally {
      clearAuth();
      router.push("/login");
    }
  };

  const NavItem = ({ item }: { item: any }) => {
    const isActive = pathname === item.href || (item.href !== "/" && pathname?.startsWith(item.href));
    
    return (
      <Link
        href={item.href}
        onClick={onNavigate}
        className={cn(
          "flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 group",
          isActive 
            ? "bg-primary text-primary-foreground shadow-md shadow-primary/20" 
            : "text-muted-foreground hover:bg-secondary/5 hover:text-secondary"
        )}
      >
        <item.icon className={cn("h-5 w-5", isActive ? "text-primary-foreground" : "text-muted-foreground group-hover:text-primary")} />
        <span className="font-medium text-sm">{item.label}</span>
      </Link>
    );
  };

  return (
    <div className="flex h-full flex-col overflow-y-auto bg-card border-r shadow-sm">
      <div className="p-6">
        <Link href="/" className="flex items-center gap-2" onClick={onNavigate}>
          <div className="bg-primary/10 p-2 rounded-xl">
            <GraduationCap className="h-6 w-6 text-primary" />
          </div>
          <span className="font-lora font-bold text-xl text-secondary tracking-tight">GCET Faculty</span>
        </Link>
      </div>
      
      <div className="flex-1 px-4 space-y-1">
        <div className="text-xs font-semibold text-muted-foreground/60 uppercase tracking-wider mb-2 ml-3">Main Menu</div>
        {routes.map((route) => (
          <NavItem key={route.href} item={route} />
        ))}
      </div>

      <div className="px-4 pb-6 space-y-1 mt-auto pt-6">
        <div className="text-xs font-semibold text-muted-foreground/60 uppercase tracking-wider mb-2 ml-3">Preferences</div>
        {bottomRoutes.map((route) => (
          <NavItem key={route.href} item={route} />
        ))}
        
        <button
          onClick={handleLogout}
          className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors text-muted-foreground hover:bg-destructive/10 hover:text-destructive group mt-4"
        >
          <LogOut className="h-5 w-5 text-muted-foreground group-hover:text-destructive" />
          <span className="font-medium text-sm">Logout</span>
        </button>
      </div>
    </div>
  );
}
