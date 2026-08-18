"use client";

import { useAuthStore } from "@/store/useAuthStore";
import { motion } from "framer-motion";
import { Calendar, ClipboardCheck, BarChart3, Clock, ArrowRight, BookOpen, Users } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { format } from "date-fns";

export default function DashboardPage() {
  const user = useAuthStore((state) => state.user);
  const [greeting, setGreeting] = useState("Welcome");
  const [date, setDate] = useState("");

  useEffect(() => {
    const hour = new Date().getHours();
    if (hour < 12) setGreeting("Good Morning");
    else if (hour < 17) setGreeting("Good Afternoon");
    else setGreeting("Good Evening");
    
    setDate(format(new Date(), "EEEE, MMMM d, yyyy"));
  }, []);

  const container = {
    hidden: { opacity: 0 },
    show: {
      opacity: 1,
      transition: { staggerChildren: 0.1 }
    }
  };

  const item = {
    hidden: { opacity: 0, y: 15 },
    show: { 
      opacity: 1, 
      y: 0, 
      transition: { type: "spring", stiffness: 300, damping: 24 }
    }
  };

  return (
    <motion.div 
      variants={container}
      initial="hidden"
      animate="show"
      className="max-w-6xl mx-auto pb-10 space-y-8 animate-in fade-in duration-500"
    >
      {/* Header Section */}
      <motion.div variants={item} className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <div>
          <div className="flex items-center gap-2 text-muted-foreground mb-2">
            <Clock className="h-4 w-4" />
            <span className="text-sm font-medium">{date}</span>
          </div>
          <h1 className="text-3xl md:text-4xl font-bold tracking-tight text-foreground">
            {greeting}, {user?.first_name || user?.name || 'Professor'}
          </h1>
          <p className="text-muted-foreground mt-2">
            Here's what's happening with your classes today.
          </p>
        </div>
        <div className="flex items-center gap-6 bg-card border rounded-xl px-5 py-3 shadow-sm shrink-0">
           <div className="flex flex-col">
             <span className="text-xs text-muted-foreground uppercase tracking-wider font-semibold mb-0.5">Department</span>
             <span className="text-sm font-medium">{user?.department || "Not assigned"}</span>
           </div>
           <div className="w-px h-8 bg-border"></div>
           <div className="flex flex-col">
             <span className="text-xs text-muted-foreground uppercase tracking-wider font-semibold mb-0.5">Employee ID</span>
             <span className="text-sm font-medium">{user?.employee_id || "N/A"}</span>
           </div>
        </div>
      </motion.div>

      {/* Main Actions Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        
        {/* Mark Attendance (Primary Action) */}
        <motion.div variants={item} className="md:col-span-2">
          <Link href="/attendance" className="block h-full group">
            <div className="h-full relative overflow-hidden rounded-2xl border bg-primary/[0.03] hover:bg-primary/[0.08] border-primary/20 transition-all duration-300 p-8 flex flex-col justify-between shadow-sm">
              <div className="flex justify-between items-start mb-8">
                <div className="p-3 bg-primary/10 rounded-xl text-primary group-hover:scale-110 transition-transform duration-300">
                  <ClipboardCheck className="h-7 w-7" />
                </div>
                <div className="p-2 bg-background rounded-full shadow-sm text-foreground opacity-0 -translate-x-4 group-hover:opacity-100 group-hover:translate-x-0 transition-all duration-300 border">
                  <ArrowRight className="h-4 w-4" />
                </div>
              </div>
              
              <div>
                <h3 className="text-2xl font-bold text-foreground mb-2">Mark Attendance</h3>
                <p className="text-muted-foreground max-w-md">
                  Start a dynamic session or manually record attendance for your current classes.
                </p>
              </div>
            </div>
          </Link>
        </motion.div>

        {/* Timetable */}
        <motion.div variants={item}>
          <Link href="/timetable" className="block h-full group">
            <div className="h-full rounded-2xl border bg-card hover:bg-muted/50 transition-all duration-300 p-8 flex flex-col justify-between shadow-sm">
              <div className="flex justify-between items-start mb-8">
                <div className="p-3 bg-blue-500/10 rounded-xl text-blue-600 group-hover:scale-110 transition-transform duration-300">
                  <Calendar className="h-6 w-6" />
                </div>
                <div className="p-2 bg-background rounded-full shadow-sm text-foreground opacity-0 -translate-x-4 group-hover:opacity-100 group-hover:translate-x-0 transition-all duration-300 border">
                  <ArrowRight className="h-4 w-4" />
                </div>
              </div>
              <div>
                <h3 className="text-xl font-bold text-foreground mb-2">Timetable</h3>
                <p className="text-sm text-muted-foreground">View your weekly schedule and upcoming lectures.</p>
              </div>
            </div>
          </Link>
        </motion.div>

        {/* Analytics */}
        <motion.div variants={item}>
          <Link href="/analytics" className="block h-full group">
            <div className="h-full rounded-2xl border bg-card hover:bg-muted/50 transition-all duration-300 p-8 flex flex-col justify-between shadow-sm">
              <div className="flex justify-between items-start mb-8">
                <div className="p-3 bg-emerald-500/10 rounded-xl text-emerald-600 group-hover:scale-110 transition-transform duration-300">
                  <BarChart3 className="h-6 w-6" />
                </div>
                <div className="p-2 bg-background rounded-full shadow-sm text-foreground opacity-0 -translate-x-4 group-hover:opacity-100 group-hover:translate-x-0 transition-all duration-300 border">
                  <ArrowRight className="h-4 w-4" />
                </div>
              </div>
              <div>
                <h3 className="text-xl font-bold text-foreground mb-2">Analytics</h3>
                <p className="text-sm text-muted-foreground">Track attendance trends and view comprehensive reports.</p>
              </div>
            </div>
          </Link>
        </motion.div>

      </div>
    </motion.div>
  );
}
