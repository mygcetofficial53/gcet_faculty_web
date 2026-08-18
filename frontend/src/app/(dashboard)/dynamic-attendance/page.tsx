"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2, Zap, Play, Square, Users, Copy, Check } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

export default function DynamicAttendancePage() {
  const [selectedCourse, setSelectedCourse] = useState<string>("");
  const [sessionActive, setSessionActive] = useState(false);
  const [code, setCode] = useState<string>("");
  const [duration, setDuration] = useState("5"); // minutes
  const [copied, setCopied] = useState(false);

  // For simulation
  const [connectedStudents, setConnectedStudents] = useState<string[]>([]);
  
  const { data: courses, isLoading } = useQuery({
    queryKey: ["attendance-courses"],
    queryFn: async () => {
      const res = await api.get("/attendance/courses");
      return res.data.data;
    },
  });

  const startSession = () => {
    // Generate a random 6 character alphanumeric code
    const newCode = Math.random().toString(36).substring(2, 8).toUpperCase();
    setCode(newCode);
    setSessionActive(true);
    
    // Simulate students joining
    let count = 0;
    const interval = setInterval(() => {
      if (count < 25) {
        setConnectedStudents(prev => [...prev, `Student ${prev.length + 1}`]);
        count++;
      } else {
        clearInterval(interval);
      }
    }, 2000);
  };

  const stopSession = () => {
    setSessionActive(false);
    setCode("");
    setConnectedStudents([]);
  };

  const copyCode = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-4xl mx-auto">
      <div>
        <h2 className="text-3xl font-bold font-lora text-secondary tracking-tight">Dynamic Attendance</h2>
        <p className="text-muted-foreground mt-1">Code-based real-time attendance tracking</p>
      </div>

      <AnimatePresence mode="wait">
        {!sessionActive ? (
          <motion.div
            key="setup"
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: 20 }}
            transition={{ duration: 0.3 }}
          >
            <Card className="shadow-sm border-border/50">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Zap className="h-5 w-5 text-primary" /> Start Session
                </CardTitle>
                <CardDescription>Select a class and generate a code for students to join.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="space-y-2">
                  <Label>Select Course</Label>
                  {isLoading ? (
                    <div className="h-10 bg-secondary/5 rounded-md flex items-center px-3 border border-border">
                      <Loader2 className="h-4 w-4 animate-spin text-muted-foreground mr-2" /> Loading courses...
                    </div>
                  ) : (
                    <Select value={selectedCourse} onValueChange={(val) => setSelectedCourse(val || "")}>
                      <SelectTrigger className="h-auto min-h-10 py-2 *:data-[slot=select-value]:line-clamp-none whitespace-normal text-left">
                        <SelectValue placeholder="Select a course..." />
                      </SelectTrigger>
                      <SelectContent className="max-w-[95vw] sm:max-w-3xl max-h-[60vh] !w-max min-w-[var(--anchor-width)]">
                        {courses?.map((c: any, idx: number) => (
                          <SelectItem key={idx} value={c.raw_value} className="whitespace-normal py-2 text-sm">{c.display_text}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                </div>

                <div className="space-y-2">
                  <Label>Duration (Minutes)</Label>
                  <Input 
                    type="number" 
                    value={duration} 
                    onChange={(e) => setDuration(e.target.value)} 
                    min="1" 
                    max="60"
                  />
                </div>

                <Button 
                  onClick={startSession} 
                  disabled={!selectedCourse}
                  className="w-full h-12 text-base"
                >
                  <Play className="mr-2 h-5 w-5" /> Start Broadcast
                </Button>
              </CardContent>
            </Card>
          </motion.div>
        ) : (
          <motion.div
            key="active"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.3 }}
            className="space-y-6"
          >
            <Card className="border-2 border-primary/50 shadow-md bg-primary/5">
              <CardContent className="p-8 text-center space-y-6">
                <div className="flex items-center justify-center space-x-2">
                  <span className="relative flex h-4 w-4">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-4 w-4 bg-emerald-500"></span>
                  </span>
                  <span className="font-medium text-emerald-600">Session Active</span>
                </div>
                
                <div>
                  <p className="text-sm text-muted-foreground mb-2 uppercase tracking-widest font-semibold">Join Code</p>
                  <div 
                    onClick={copyCode}
                    className="flex items-center justify-center gap-4 bg-white dark:bg-black rounded-2xl py-4 px-8 border shadow-sm mx-auto w-fit cursor-pointer hover:border-primary/50 transition-colors"
                  >
                    <span className="text-6xl font-bold font-mono tracking-[0.2em] text-secondary">
                      {code}
                    </span>
                    <Button variant="ghost" size="icon" className="shrink-0 text-muted-foreground hover:text-primary">
                      {copied ? <Check className="h-6 w-6 text-emerald-500" /> : <Copy className="h-6 w-6" />}
                    </Button>
                  </div>
                </div>

                <p className="text-muted-foreground">
                  Ask students to enter this code in their GCET app.
                </p>

                <div className="pt-4 flex justify-center">
                  <Button variant="destructive" onClick={stopSession} className="h-12 px-8">
                    <Square className="mr-2 h-5 w-5 fill-current" /> End Session
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-lg">Connected Students</CardTitle>
                <div className="flex items-center text-emerald-600 font-bold bg-emerald-100 px-3 py-1 rounded-full">
                  <Users className="mr-2 h-4 w-4" /> {connectedStudents.length}
                </div>
              </CardHeader>
              <CardContent>
                {connectedStudents.length === 0 ? (
                  <div className="py-8 text-center text-muted-foreground">
                    Waiting for students to join...
                  </div>
                ) : (
                  <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3 mt-4">
                    <AnimatePresence>
                      {connectedStudents.map((s, i) => (
                        <motion.div
                          key={i}
                          initial={{ opacity: 0, scale: 0.8 }}
                          animate={{ opacity: 1, scale: 1 }}
                          className="bg-secondary/5 px-3 py-2 rounded-lg text-sm font-medium text-secondary flex items-center gap-2 border"
                        >
                          <div className="h-2 w-2 rounded-full bg-emerald-500 shrink-0" />
                          <span className="truncate">{s}</span>
                        </motion.div>
                      ))}
                    </AnimatePresence>
                  </div>
                )}
              </CardContent>
            </Card>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
