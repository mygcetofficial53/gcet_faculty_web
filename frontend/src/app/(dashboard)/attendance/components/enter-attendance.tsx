"use client";

import { useState, useRef, useEffect, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { format } from "date-fns";
import { cn } from "@/lib/utils";
import { Loader2, Users, Save, CheckCircle2, Calendar as CalendarIcon, Radio, Info } from "lucide-react";
import { DynamicAttendance } from "./dynamic-attendance";

export default function EnterAttendance() {
  const [selectedCourse, setSelectedCourse] = useState<string>("");
  const [date, setDate] = useState<string>(new Date().toISOString().split("T")[0]);
  const [topic, setTopic] = useState<string>("");
  const [extraLecture, setExtraLecture] = useState<boolean>(false);
  const [lecNo, setLecNo] = useState<string>("1");
  const [lectureType, setLectureType] = useState<string>("L");
  const [students, setStudents] = useState<any[]>([]);
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [success, setSuccess] = useState(false);
  
  const d2dStartIndex = useMemo(() => {
    if (students.length < 2) return -1;
    const firstValidIndex = students.findIndex(s => s.enrollment && s.enrollment.length >= 3);
    if (firstValidIndex === -1) return -1;
    
    const firstPrefix = students[firstValidIndex].enrollment.substring(0, 3);
    for (let i = firstValidIndex + 1; i < students.length; i++) {
      const currentEnr = students[i].enrollment;
      if (currentEnr && currentEnr.length >= 3) {
        if (currentEnr.substring(0, 3) !== firstPrefix) {
          return i + 1; // Return Sr. No (1-indexed)
        }
      }
    }
    return -1;
  }, [students]);

  // Navigation, Input Modes & Filters
  const [focusedRowIndex, setFocusedRowIndex] = useState<number>(0);
  const [entryMode, setEntryMode] = useState<"list" | "quick">("list");
  const [filterMode, setFilterMode] = useState<"all" | "present" | "absent">("all");
  const [quickInput, setQuickInput] = useState<string>("");
  const tableRef = useRef<HTMLDivElement>(null);

  // Keyboard navigation effect
  useEffect(() => {
    if (step === 2 && entryMode === "list" && tableRef.current) {
      tableRef.current.focus();
    }
  }, [step, entryMode]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (entryMode !== "list") return;
    
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setFocusedRowIndex((prev) => {
        const next = Math.min(prev + 1, students.length - 1);
        document.getElementById(`student-row-${next}`)?.scrollIntoView({ block: "nearest" });
        return next;
      });
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setFocusedRowIndex((prev) => {
        const next = Math.max(prev - 1, 0);
        document.getElementById(`student-row-${next}`)?.scrollIntoView({ block: "nearest" });
        return next;
      });
    } else if (e.key === " ") {
      e.preventDefault();
      toggleStudent(focusedRowIndex);
    }
  };

  const handleQuickInputApply = () => {
    // parse input like "1, 4, 12-15" (assuming these are last few digits of enrollment or Sr No)
    // Actually, let's just make it simple: "Mark Absent by Sr. No."
    // Because enrollment numbers are long.
    const parts = quickInput.split(",").map(s => s.trim()).filter(Boolean);
    const indicesToMarkAbsent = new Set<number>();
    
    parts.forEach(part => {
      const targetNumbers: { num: number, isD2DTarget: boolean }[] = [];
      let str = part.toLowerCase();
      let isD2DTarget = false;
      
      if (str.startsWith('d')) {
        isD2DTarget = true;
        str = str.replace(/d/g, ''); // parse "d1-d5" -> "1-5"
      }

      if (str.includes("-")) {
        const [start, end] = str.split("-").map(n => parseInt(n, 10));
        if (!isNaN(start) && !isNaN(end)) {
          for (let i = start; i <= end; i++) {
            targetNumbers.push({ num: i, isD2DTarget });
          }
        }
      } else {
        const num = parseInt(str, 10);
        if (!isNaN(num)) {
          targetNumbers.push({ num, isD2DTarget });
        }
      }

      targetNumbers.forEach(({ num, isD2DTarget }) => {
        const matchingIndices: number[] = [];
        students.forEach((s, idx) => {
          if (s.enrollment) {
            // Extract the last 3 digits of enrollment (the standard GTU Roll No)
            // e.g., "12402080501004" -> "004" -> 4
            const rollNoStr = s.enrollment.slice(-3);
            const rollNo = parseInt(rollNoStr, 10);
            if (!isNaN(rollNo) && rollNo === num) {
              const studentIsD2D = d2dStartIndex !== -1 && idx >= d2dStartIndex - 1;
              
              if (isD2DTarget) {
                // If they typed "d4", ONLY match D2D students
                if (studentIsD2D) matchingIndices.push(idx);
              } else {
                // If they typed "4", ONLY match Regular students (if we know who they are)
                if (d2dStartIndex !== -1) {
                  if (!studentIsD2D) matchingIndices.push(idx);
                } else {
                  // If we don't know who D2D students are, just match anyone
                  matchingIndices.push(idx);
                }
              }
            }
          }
        });

        if (matchingIndices.length > 0) {
          // Found student(s) with this true Roll No. 
          indicesToMarkAbsent.add(matchingIndices[0]);
        } else {
          // Fallback to Sr. No if no student matched this Roll No
          if (num > 0 && num <= students.length) {
            indicesToMarkAbsent.add(num - 1);
          }
        }
      });
    });

    const newStudents = students.map((s, idx) => ({
      ...s,
      is_present: !indicesToMarkAbsent.has(idx)
    }));
    setStudents(newStudents);
  };


  const queryClient = useQueryClient();

  // 1. Fetch Courses
  const { data: courses, isLoading: loadingCourses } = useQuery({
    queryKey: ["attendance-courses"],
    queryFn: async () => {
      const res = await api.get("/attendance/courses");
      return res.data.data;
    },
  });

  // 2. Fetch Students Mutation
  const fetchStudents = useMutation({
    mutationFn: async () => {
      const course = courses.find((c: any) => c.raw_value === selectedCourse);
      const res = await api.post("/attendance/students", {
        course_code: course.course_code,
        date: date,
        by_lib_id: false,
        is_edit: false,
        metadata: course.metadata,
        option_index: course.option_index
      });
      return res.data.data;
    },
    onSuccess: (data) => {
      setStudents(data);
      setStep(2);
    }
  });

  // 3. Submit Attendance Mutation
  const submitAttendance = useMutation({
    mutationFn: async () => {
      const course = courses.find((c: any) => c.raw_value === selectedCourse);
      const res = await api.post("/attendance/enter", {
        course_code: course.raw_value,
        date: date,
        time_slot: "", 
        topic: topic,
        type: lectureType,
        students: students,
        extra_lecture: extraLecture,
        lec_no: extraLecture ? lecNo : "1",
        metadata: course.metadata,
        option_index: course.option_index
      });
      return res.data;
    },
    onSuccess: () => {
      setSuccess(true);
      queryClient.invalidateQueries({ queryKey: ["attendance-sheets"] });
      queryClient.invalidateQueries({ queryKey: ["attendance-courses"] });
      setTimeout(() => {
        setStep(1);
        setSuccess(false);
        setStudents([]);
        setSelectedCourse("");
      }, 3000);
    }
  });

  const toggleStudent = (index: number) => {
    const newStudents = [...students];
    newStudents[index].is_present = !newStudents[index].is_present;
    setStudents(newStudents);
  };

  const toggleAll = (present: boolean) => {
    setStudents(students.map(s => ({ ...s, is_present: present })));
  };

  if (loadingCourses) {
    return <div className="flex justify-center p-8"><Loader2 className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  if (success) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center animate-in zoom-in duration-300">
        <div className="h-20 w-20 bg-emerald-100 rounded-full flex items-center justify-center mb-6">
          <CheckCircle2 className="h-10 w-10 text-emerald-600" />
        </div>
        <h3 className="text-2xl font-bold text-secondary mb-2">Attendance Submitted!</h3>
        <p className="text-muted-foreground">The attendance has been successfully recorded in the GMS portal.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {step === 1 && (
        <div className="space-y-4 max-w-2xl animate-in fade-in duration-300">
          <div className="space-y-2">
            <Label>Select Course / Batch</Label>
            <Select value={selectedCourse} onValueChange={(val) => setSelectedCourse(val || "")}>
              <SelectTrigger className="h-auto min-h-12 py-3 *:data-[slot=select-value]:line-clamp-none whitespace-normal text-left">
                <SelectValue placeholder="Select a course to mark attendance">
                  {selectedCourse && courses ? courses.find((c: any) => String(c.raw_value) === selectedCourse)?.display_text || selectedCourse : undefined}
                </SelectValue>
              </SelectTrigger>
              <SelectContent className="max-w-[95vw] sm:max-w-3xl max-h-[60vh] !w-max min-w-[var(--anchor-width)]">
                {courses?.map((course: any, idx: number) => (
                  <SelectItem key={idx} value={String(course.raw_value)} className="whitespace-normal py-3 text-sm">
                    {course.display_text}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="space-y-2 flex flex-col">
              <Label>Date</Label>
              <Popover>
                <PopoverTrigger
                  render={
                    <Button
                      variant={"outline"}
                      className={cn(
                        "w-full h-12 justify-start text-left font-normal border-input bg-background hover:bg-accent hover:text-accent-foreground",
                        !date && "text-muted-foreground"
                      )}
                    >
                      <CalendarIcon className="mr-2 h-4 w-4" />
                      {date ? format(new Date(date), "PPP") : <span>Pick a date</span>}
                    </Button>
                  }
                />
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={date ? new Date(date) : undefined}
                    onSelect={(d) => d && setDate(format(d, "yyyy-MM-dd"))}
                  />
                </PopoverContent>
              </Popover>
              {selectedCourse && courses?.find((c: any) => c.raw_value === selectedCourse)?.last_entered_dates?.length > 0 && (
                <div className="text-xs text-muted-foreground mt-1">
                  Last marked: {courses.find((c: any) => c.raw_value === selectedCourse).last_entered_dates.join(", ")}
                </div>
              )}
            </div>
            <div className="space-y-2">
              <Label>Topic Covered</Label>
              <Input 
                value={topic} 
                onChange={(e) => setTopic(e.target.value)} 
                placeholder="e.g. Introduction to Logic"
                className="h-12"
              />
            </div>
          </div>

          <div className="space-y-3">
            <Label>Type</Label>
            <div className="flex gap-4">
              <label className="flex items-center space-x-2 cursor-pointer">
                <input 
                  type="radio" 
                  name="type" 
                  value="L" 
                  checked={lectureType === "L"} 
                  onChange={(e) => setLectureType(e.target.value)} 
                  className="h-4 w-4 text-primary focus:ring-primary border-gray-300"
                />
                <span className="text-sm font-medium">Lecture</span>
              </label>
              <label className="flex items-center space-x-2 cursor-pointer">
                <input      
                  type="radio" 
                  name="type" 
                  value="P" 
                  checked={lectureType === "P"} 
                  onChange={(e) => setLectureType(e.target.value)}
                  className="h-4 w-4 text-primary focus:ring-primary border-gray-300"
                />
                <span className="text-sm font-medium">Practical</span>
              </label>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 border-t pt-4 mt-2">
            <div className="space-y-2">
              <div className="flex items-center h-12 space-x-2">
                <Checkbox 
                  id="extra-lecture" 
                  checked={extraLecture}
                  onCheckedChange={(checked) => setExtraLecture(checked as boolean)}
                />
                <Label htmlFor="extra-lecture" className="cursor-pointer">Extra Lecture?</Label>
              </div>
            </div>
            {extraLecture && (
              <div className="space-y-2">
                <Label>Lecturer No</Label>
                <Input 
                  type="number" 
                  value={lecNo} 
                  onChange={(e) => setLecNo(e.target.value)} 
                  placeholder="1"
                  className="h-12"
                />
              </div>
            )}
          </div>

          <Button 
            className="w-full h-12 mt-4 text-base" 
            onClick={() => fetchStudents.mutate()}
            disabled={!selectedCourse || fetchStudents.isPending}
          >
            {fetchStudents.isPending ? <Loader2 className="mr-2 h-5 w-5 animate-spin" /> : <Users className="mr-2 h-5 w-5" />}
            Fetch Student List
          </Button>
          {fetchStudents.isError && (
            <p className="text-sm text-destructive mt-2 text-center">Failed to fetch students. Please try again.</p>
          )}
        </div>
      )}

      {step === 2 && (
        <div className="animate-in slide-in-from-right-4 duration-300 space-y-6">
            <div className="bg-blue-50/50 border border-blue-200 rounded-xl p-4 flex gap-3 items-start relative">
              <Info className="h-5 w-5 text-blue-500 mt-0.5 flex-shrink-0" />
              <div className="space-y-1">
                <h4 className="font-semibold text-blue-800">New: Faster Attendance Entry!</h4>
                <ul className="text-sm text-blue-700/90 list-disc list-inside space-y-1 ml-1">
                  <li><strong>Keyboard Navigation:</strong> In List View, use <kbd className="bg-white border shadow-sm rounded px-1.5 py-0.5 text-xs font-sans text-blue-800">↑</kbd> <kbd className="bg-white border shadow-sm rounded px-1.5 py-0.5 text-xs font-sans text-blue-800">↓</kbd> arrows to navigate students and <kbd className="bg-white border shadow-sm rounded px-1.5 py-0.5 text-xs font-sans text-blue-800">Space</kbd> to mark present/absent.</li>
                  <li>
                    <strong>Quick Input:</strong> Switch modes to type the absent numbers. You can type their <strong>Roll Number</strong> (the last 3 digits of enrollment) OR their Sr. No. 
                    <br/><span className="text-blue-600">💡 <strong>Pro Tip for Detained Students:</strong> If a student is detained and the Sr. Nos shift on your printed sheet, don't worry! Just type the student's <strong>Roll Number</strong>! For example, if you type <code className="bg-white px-1.5 py-0.5 rounded text-blue-800">4</code>, it finds the Regular student whose enrollment ends in `004`! 
                    <br/>If you want to mark a <strong>D2D student</strong> absent, simply prefix it with <strong>d</strong>! For example, type <code className="bg-white px-1.5 py-0.5 rounded text-blue-800">d4</code> or <code className="bg-white px-1.5 py-0.5 rounded text-blue-800">d1-d5</code> to instantly find the D2D students!</span>
                  </li>
                </ul>
              </div>
            </div>

          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-secondary/5 p-4 rounded-xl border">
            <div>
              <h3 className="font-semibold text-secondary">{courses?.find((c: any) => c.raw_value === selectedCourse)?.display_text}</h3>
              <p className="text-sm text-muted-foreground mt-1">Date: {date} • Topic: {topic}</p>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => toggleAll(true)}>Mark All Present</Button>
              <Button variant="outline" size="sm" onClick={() => toggleAll(false)}>Mark All Absent</Button>
            </div>
          </div>

          <div className="flex gap-2 p-1 bg-muted rounded-lg w-max mb-4">
            <button 
              className={`px-4 py-2 text-sm font-medium rounded-md transition-all ${entryMode === 'list' ? 'bg-background shadow text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
              onClick={() => setEntryMode('list')}
            >
              List View (Checkboxes)
            </button>
            <button 
              className={`px-4 py-2 text-sm font-medium rounded-md transition-all ${entryMode === 'quick' ? 'bg-background shadow text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
              onClick={() => setEntryMode('quick')}
            >
              Quick Input
            </button>
          </div>

          {entryMode === 'quick' && (
            <div className="bg-card border rounded-xl p-6 space-y-4 mb-4 animate-in fade-in slide-in-from-top-2">
              <div>
                <h4 className="text-lg font-medium">Quick Absent Entry</h4>
                <p className="text-sm text-muted-foreground mt-1">
                  Type the <strong>Sr. No.</strong> of the absent students to quickly mark them. Everyone else will be marked present.
                  <br/>Example: <code className="bg-muted px-1.5 py-0.5 rounded text-primary">1, 4, 12-15</code>
                </p>
              </div>
              <div className="flex gap-3">
                <Input 
                  value={quickInput}
                  onChange={(e) => setQuickInput(e.target.value)}
                  placeholder="e.g., 1, 4, 12-15" 
                  className="max-w-md h-12 text-lg"
                />
                <Button onClick={handleQuickInputApply} className="h-12">Apply to List</Button>
              </div>
            </div>
          )}

          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-2">
            <div className="text-sm font-medium text-muted-foreground">
              Showing: {filterMode === 'all' ? 'All Students' : filterMode === 'present' ? 'Present Students Only' : 'Absent Students Only'}
            </div>
            <div className="flex bg-muted p-1 rounded-lg w-max">
              <button 
                className={`px-3 py-1.5 text-xs font-medium rounded-md transition-all ${filterMode === 'all' ? 'bg-background shadow text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
                onClick={() => setFilterMode('all')}
              >
                All
              </button>
              <button 
                className={`px-3 py-1.5 text-xs font-medium rounded-md transition-all ${filterMode === 'present' ? 'bg-background shadow text-foreground text-emerald-600' : 'text-muted-foreground hover:text-foreground'}`}
                onClick={() => setFilterMode('present')}
              >
                Present
              </button>
              <button 
                className={`px-3 py-1.5 text-xs font-medium rounded-md transition-all ${filterMode === 'absent' ? 'bg-background shadow text-foreground text-destructive' : 'text-muted-foreground hover:text-foreground'}`}
                onClick={() => setFilterMode('absent')}
              >
                Absent
              </button>
            </div>
          </div>

          <div 
            className="border rounded-xl overflow-hidden bg-card focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 transition-shadow"
            tabIndex={entryMode === "list" ? 0 : -1}
            ref={tableRef}
            onKeyDown={handleKeyDown}
          >
            <div className="max-h-[500px] overflow-auto custom-scrollbar">
              <table className="w-full text-sm text-left">
                <thead className="text-xs uppercase bg-muted/50 sticky top-0 z-10">
                  <tr>
                    <th className="px-4 py-3 w-16 text-center">Sr.</th>
                    <th className="px-4 py-3">Enrollment No</th>
                    <th className="px-4 py-3">Student Name</th>
                    <th className="px-4 py-3 text-center w-28">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {students.map((student, idx) => {
                    if (filterMode === 'present' && !student.is_present) return null;
                    if (filterMode === 'absent' && student.is_present) return null;

                    return (
                      <tr 
                        key={student.enrollment} 
                        id={`student-row-${idx}`}
                        className={`cursor-pointer transition-colors ${
                          focusedRowIndex === idx && entryMode === 'list' && filterMode === 'all'
                            ? 'bg-primary/10 border-l-4 border-primary' 
                            : !student.is_present 
                              ? 'bg-destructive/5 hover:bg-destructive/10' 
                              : 'hover:bg-muted/30'
                        }`}
                        onClick={() => {
                          if (filterMode === 'all') setFocusedRowIndex(idx);
                          toggleStudent(idx);
                        }}
                      >
                        <td className="px-4 py-3 text-center text-muted-foreground">{idx + 1}</td>
                        <td className="px-4 py-3 font-medium">{student.enrollment}</td>
                        <td className="px-4 py-3">{student.name}</td>
                        <td className="px-4 py-3 text-center" onClick={(e) => e.stopPropagation()}>
                          <Checkbox 
                            checked={student.is_present}
                            onCheckedChange={() => toggleStudent(idx)}
                            className={student.is_present ? "data-[state=checked]:bg-emerald-500 data-[state=checked]:border-emerald-500 h-5 w-5" : "h-5 w-5 border-destructive"}
                          />
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>

          <div className="flex flex-col sm:flex-row justify-between sm:items-center gap-4 bg-card p-4 rounded-xl border shadow-sm sticky bottom-0 z-20">
            <div className="flex items-center gap-6">
              <div className="text-sm">
                <span className="text-muted-foreground">Present:</span>
                <span className="ml-2 font-bold text-emerald-600 text-lg">{students.filter(s => s.is_present).length}</span>
              </div>
              <div className="text-sm">
                <span className="text-muted-foreground">Absent:</span>
                <span className="ml-2 font-bold text-destructive text-lg">{students.filter(s => !s.is_present).length}</span>
              </div>
            </div>
            <div className="flex gap-3">
              <Button variant="ghost" onClick={() => setStep(1)} disabled={submitAttendance.isPending}>Back</Button>
              <Button onClick={() => submitAttendance.mutate()} disabled={submitAttendance.isPending} className="bg-primary hover:bg-primary/90">
                {submitAttendance.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                Submit Attendance
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
