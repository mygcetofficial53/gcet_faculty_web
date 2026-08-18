"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Clock, MapPin, BookOpen, Users, Plus, MoreVertical, Trash2, EyeOff, RotateCcw } from "lucide-react";

interface TimetableEntry {
  day: string;
  time: string;
  subject: string;
  type: string;
  room: string;
  batch: string;
  classroom: string;
  is_custom?: boolean;
}

const DAYS = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

function formatTime12h(time24: string) {
  if (!time24) return "";
  const [h, m] = time24.split(":");
  let hour = parseInt(h, 10);
  const ampm = hour >= 12 ? "PM" : "AM";
  hour = hour % 12;
  hour = hour ? hour : 12;
  return `${hour}:${m} ${ampm}`;
}

export default function TimetablePage() {
  const queryClient = useQueryClient();
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [startTime, setStartTime] = useState("10:30");
  const [endTime, setEndTime] = useState("11:30");
  const [newEntry, setNewEntry] = useState({
    day: "Monday",
    subject: "",
    type: "L",
    room: "",
    batch: "All",
  });

  const { data: courses, isLoading: loadingCourses } = useQuery({
    queryKey: ["attendance-courses"],
    queryFn: async () => {
      const res = await api.get("/attendance/courses");
      return res.data.data;
    },
  });

  const { data, isLoading, error } = useQuery({
    queryKey: ["timetable"],
    queryFn: async () => {
      const res = await api.get("/timetable");
      if (!res.data.success) throw new Error(res.data.error || "Failed to load timetable");
      return res.data.data as TimetableEntry[];
    },
  });

  const getEntriesForDay = (day: string) => {
    if (!data) return [];
    return data.filter((entry) => entry.day.toLowerCase() === day.toLowerCase());
  };

  const addMutation = useMutation({
    mutationFn: async (entry: any) => {
      const formattedTime = `${formatTime12h(startTime)} - ${formatTime12h(endTime)}`;
      const res = await api.post("/timetable/custom", {
        day: entry.day,
        time: formattedTime,
        subject: entry.subject,
        class_type: entry.type,
        room: entry.room,
        batch: entry.batch,
      });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["timetable"] });
      setIsAddDialogOpen(false);
      setNewEntry({ day: "Monday", subject: "", type: "L", room: "", batch: "All" });
      setStartTime("10:30");
      setEndTime("11:30");
    }
  });

  const deleteMutation = useMutation({
    mutationFn: async (entry: TimetableEntry) => {
      const res = await api.delete("/timetable/custom", { data: { day: entry.day, time: entry.time, subject: entry.subject } });
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["timetable"] })
  });

  const hideMutation = useMutation({
    mutationFn: async (entry: TimetableEntry) => {
      const res = await api.post("/timetable/hide", { day: entry.day, time: entry.time, subject: entry.subject });
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["timetable"] })
  });

  const resetMutation = useMutation({
    mutationFn: async () => {
      const res = await api.post("/timetable/reset");
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["timetable"] })
  });

  const today = new Date().toLocaleDateString("en-US", { weekday: "long" });
  const defaultTab = DAYS.includes(today) ? today : "Monday";

  if (error) {
    return (
      <div className="p-4 rounded-lg bg-destructive/10 text-destructive text-center">
        Error loading timetable: {(error as Error).message}
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h2 className="text-3xl font-bold font-lora text-secondary tracking-tight">Timetable</h2>
          <p className="text-muted-foreground mt-1">Your weekly teaching schedule</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => resetMutation.mutate()} disabled={resetMutation.isPending}>
            <RotateCcw className="mr-2 h-4 w-4" /> Reset 
          </Button>
          <Button onClick={() => setIsAddDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" /> Add Class
          </Button>
        </div>
      </div>

      <Card className="shadow-sm border-border/50">
        <Tabs defaultValue={defaultTab} className="w-full">
          <CardHeader className="pb-3 border-b px-4 sm:px-6">
            <div className="w-full overflow-x-auto hide-scrollbar pb-1 -mb-1">
              <TabsList className="inline-flex w-max min-w-full justify-start h-auto p-1 bg-secondary/5">
                {DAYS.map((day) => (
                  <TabsTrigger 
                    key={day} 
                    value={day}
                    className="flex-shrink-0 px-4 py-2 text-sm font-medium data-[state=active]:bg-white data-[state=active]:text-primary data-[state=active]:shadow-sm rounded-md transition-all"
                  >
                    {day}
                  </TabsTrigger>
                ))}
              </TabsList>
            </div>
          </CardHeader>

          <CardContent className="pt-6">
            {isLoading ? (
              <div className="space-y-4">
                {[1, 2, 3].map((i) => (
                  <Skeleton key={i} className="h-24 w-full rounded-xl" />
                ))}
              </div>
            ) : (
              DAYS.map((day) => (
                <TabsContent key={day} value={day} className="m-0 space-y-4">
                  {getEntriesForDay(day).length === 0 ? (
                    <div className="text-center py-12 bg-secondary/5 rounded-xl border border-dashed border-border">
                      <p className="text-muted-foreground">No classes scheduled for {day}</p>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      {getEntriesForDay(day).map((entry, idx) => (
                        <Card key={idx} className="overflow-hidden border-l-4 border-l-primary shadow-sm hover:shadow-md transition-shadow relative">
                          <CardContent className="p-5">
                            <div className="flex justify-between items-start mb-4">
                              <div className="pr-6">
                                <div className="font-bold text-lg text-secondary line-clamp-2" title={entry.subject}>
                                  {entry.subject}
                                </div>
                                <div className="flex items-center gap-2 mt-1">
                                  <div className="px-2.5 py-1 rounded-full bg-primary/10 text-primary text-xs font-semibold whitespace-nowrap">
                                    {entry.type}
                                  </div>
                                  {entry.is_custom && (
                                    <div className="px-2.5 py-1 rounded-full bg-emerald-500/10 text-emerald-600 text-xs font-semibold whitespace-nowrap">
                                      Custom
                                    </div>
                                  )}
                                </div>
                              </div>
                              <DropdownMenu>
                                <DropdownMenuTrigger
                                  render={
                                    <Button variant="ghost" className="h-8 w-8 p-0 absolute top-4 right-4 text-muted-foreground hover:text-foreground">
                                      <MoreVertical className="h-4 w-4" />
                                    </Button>
                                  }
                                />
                                <DropdownMenuContent align="end">
                                  {entry.is_custom ? (
                                    <DropdownMenuItem className="text-destructive focus:text-destructive cursor-pointer" onClick={() => deleteMutation.mutate(entry)}>
                                      <Trash2 className="h-4 w-4 mr-2" /> Delete Class
                                    </DropdownMenuItem>
                                  ) : (
                                    <DropdownMenuItem className="text-muted-foreground focus:text-foreground cursor-pointer" onClick={() => hideMutation.mutate(entry)}>
                                      <EyeOff className="h-4 w-4 mr-2" /> Hide Class
                                    </DropdownMenuItem>
                                  )}
                                </DropdownMenuContent>
                              </DropdownMenu>
                            </div>
                            
                            <div className="space-y-2.5">
                              <div className="flex items-center text-sm text-muted-foreground">
                                <Clock className="h-4 w-4 mr-2.5 text-secondary/60" />
                                {entry.time}
                              </div>
                              
                              <div className="flex items-center text-sm text-muted-foreground">
                                <Users className="h-4 w-4 mr-2.5 text-secondary/60" />
                                {entry.batch || "All"}
                              </div>
                              
                              <div className="flex items-center text-sm text-muted-foreground">
                                <MapPin className="h-4 w-4 mr-2.5 text-secondary/60" />
                                <span className="truncate" title={entry.room || entry.classroom}>
                                  {entry.room || entry.classroom || "N/A"}
                                </span>
                              </div>
                            </div>
                          </CardContent>
                        </Card>
                      ))}
                    </div>
                  )}
                </TabsContent>
              ))
            )}
          </CardContent>
        </Tabs>
      </Card>

      <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Add Custom Class</DialogTitle>
            <DialogDescription>
              Add a new class to your timetable. This will sync with your mobile app.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label>Day</Label>
              <Select value={newEntry.day} onValueChange={(val) => setNewEntry({...newEntry, day: val || ""})}>
                <SelectTrigger><SelectValue placeholder="Select day" /></SelectTrigger>
                <SelectContent>
                  {DAYS.map(d => <SelectItem key={d} value={d}>{d}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label>Time</Label>
              <div className="flex items-center gap-2">
                <Input type="time" value={startTime} onChange={(e) => setStartTime(e.target.value)} className="w-full h-10" />
                <span className="text-muted-foreground">-</span>
                <Input type="time" value={endTime} onChange={(e) => setEndTime(e.target.value)} className="w-full h-10" />
              </div>
            </div>
            <div className="grid gap-2">
              <Label>Subject</Label>
              {loadingCourses ? (
                <Skeleton className="h-10 w-full" />
              ) : (
                <Select value={newEntry.subject} onValueChange={(val) => setNewEntry({...newEntry, subject: val || ""})}>
                  <SelectTrigger className="h-auto min-h-10 py-2 *:data-[slot=select-value]:line-clamp-none whitespace-normal text-left">
                    <SelectValue placeholder="Select a subject" />
                  </SelectTrigger>
                  <SelectContent className="max-w-[95vw] sm:max-w-3xl max-h-[60vh] !w-max min-w-[var(--radix-select-trigger-width)]">
                    {courses?.map((course: any, idx: number) => (
                      <SelectItem key={idx} value={course.display_text} className="whitespace-normal py-2 text-sm pr-6">
                        {course.display_text}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </div>
            <div className="grid gap-2">
              <Label>Type</Label>
              <Select value={newEntry.type} onValueChange={(val) => setNewEntry({...newEntry, type: val || ""})}>
                <SelectTrigger><SelectValue placeholder="Select type" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="L">Lecture (L)</SelectItem>
                  <SelectItem value="P">Practical (P)</SelectItem>
                  <SelectItem value="T">Tutorial (T)</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label>Room / Lab</Label>
                <Input value={newEntry.room} onChange={(e) => setNewEntry({...newEntry, room: e.target.value})} placeholder="e.g. A302" />
              </div>
              <div className="grid gap-2">
                <Label>Batch</Label>
                <Input value={newEntry.batch} onChange={(e) => setNewEntry({...newEntry, batch: e.target.value})} placeholder="e.g. B1" />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsAddDialogOpen(false)}>Cancel</Button>
            <Button onClick={() => addMutation.mutate(newEntry)} disabled={addMutation.isPending || !newEntry.subject || !startTime || !endTime}>
              {addMutation.isPending ? "Adding..." : "Add Class"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
