"use client";

import { useAuthStore } from "@/store/useAuthStore";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Menu, Bell } from "lucide-react";
import { Sidebar } from "./sidebar";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useState } from "react";
import { usePathname } from "next/navigation";

export function Navbar() {
  const user = useAuthStore((state) => state.user);
  const [open, setOpen] = useState(false);
  const pathname = usePathname();

  const getPageTitle = () => {
    switch (pathname) {
      case "/": return "Dashboard";
      case "/timetable": return "Timetable";
      case "/attendance": return "Attendance";
      case "/dynamic-attendance": return "Dynamic Attendance";
      case "/analytics": return "Analytics";
      case "/profile": return "Profile";
      case "/settings": return "Settings";
      case "/support": return "Support & Feedback";
      case "/support-developer": return "Support Developer";
      case "/about": return "About";
      default: 
        if (pathname?.startsWith("/attendance/")) return "Manage Attendance";
        return "GCET Faculty Portal";
    }
  };

  const getInitials = (name?: string) => {
    if (!name) return "F";
    const parts = name.split(" ");
    if (parts.length >= 2) {
      return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
    }
    return name.substring(0, 2).toUpperCase();
  };

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center gap-4 border-b bg-background/80 backdrop-blur-md px-4 sm:px-6 shadow-sm">
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger className="md:hidden shrink-0 inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground h-10 w-10">
          <Menu className="h-6 w-6" />
          <span className="sr-only">Toggle navigation menu</span>
        </SheetTrigger>
        <SheetContent side="left" className="p-0 w-72">
          <Sidebar onNavigate={() => setOpen(false)} />
        </SheetContent>
      </Sheet>
      
      <div className="flex flex-1 items-center justify-between">
        <h1 className="text-xl font-lora font-bold text-secondary truncate">
          {getPageTitle()}
        </h1>
        
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" className="text-muted-foreground relative">
            <Bell className="h-5 w-5" />
            <span className="absolute top-2 right-2.5 h-2 w-2 rounded-full bg-destructive border-2 border-background"></span>
          </Button>
          
          <div className="flex items-center gap-3 pl-4 border-l">
            <div className="hidden sm:flex flex-col items-end text-sm">
              <span className="font-medium leading-none text-secondary">{user?.name || 'Faculty Member'}</span>
              <span className="text-xs text-muted-foreground mt-1">{user?.department || 'GCET'}</span>
            </div>
            <Avatar className="h-9 w-9 border-2 border-primary/20">
              <AvatarImage src={`https://api.dicebear.com/7.x/initials/svg?seed=${user?.name || 'GCET'}&backgroundColor=D35D27&textColor=ffffff`} alt={user?.name || ''} />
              <AvatarFallback className="bg-primary/10 text-primary font-medium">
                {getInitials(user?.name)}
              </AvatarFallback>
            </Avatar>
          </div>
        </div>
      </div>
    </header>
  );
}
