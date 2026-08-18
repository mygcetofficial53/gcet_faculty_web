"use client";

import { useState } from "react";
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
import { Loader2, Radio, Calendar as CalendarIcon, CheckCircle2 } from "lucide-react";
import { DynamicAttendance } from "./dynamic-attendance";

export default function DynamicSessionTab() {
  const [selectedCourse, setSelectedCourse] = useState<string>("");
  const [date, setDate] = useState<string>(new Date().toISOString().split("T")[0]);
  const [topic, setTopic] = useState<string>("");
  const [extraLecture, setExtraLecture] = useState<boolean>(false);
  const [lecNo, setLecNo] = useState<string>("1");
  const [lectureType, setLectureType] = useState<string>("L");
  const [students, setStudents] = useState<any[]>([]);
  const [step, setStep] = useState<1 | 2>(1);
  const [success, setSuccess] = useState(false);
  
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
        course_code: course.raw_value,
        date: date,
        time_slot: "", 
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

  if (loadingCourses) {
    return <div className="flex justify-center p-8"><Loader2 className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  if (success) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center animate-in zoom-in duration-300">
        <div className="h-20 w-20 bg-emerald-100 rounded-full flex items-center justify-center mb-6">
          <CheckCircle2 className="h-10 w-10 text-emerald-600" />
        </div>
        <h3 className="text-2xl font-bold text-secondary mb-2">Session Completed!</h3>
        <p className="text-muted-foreground">The dynamic attendance session has been successfully recorded in the GMS portal.</p>
        <Button className="mt-8" onClick={() => setSuccess(false)}>Start Another Session</Button>
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
                <SelectValue placeholder="Select a course to start a dynamic session" />
              </SelectTrigger>
              <SelectContent className="max-w-[95vw] sm:max-w-3xl max-h-[60vh] !w-max min-w-[var(--anchor-width)]">
                {courses?.map((course: any, idx: number) => (
                  <SelectItem key={idx} value={course.raw_value} className="whitespace-normal py-3 text-sm">
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
                  id="extra-lecture-dynamic" 
                  checked={extraLecture}
                  onCheckedChange={(checked) => setExtraLecture(checked as boolean)}
                />
                <Label htmlFor="extra-lecture-dynamic" className="cursor-pointer">Extra Lecture?</Label>
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
            className="w-full h-12 mt-4 text-base font-bold text-emerald-600 border-emerald-600 hover:bg-emerald-50" 
            variant="outline"
            onClick={() => fetchStudents.mutate()}
            disabled={!selectedCourse || fetchStudents.isPending}
          >
            {fetchStudents.isPending ? <Loader2 className="mr-2 h-5 w-5 animate-spin" /> : <Radio className="mr-2 h-5 w-5" />}
            Continue to Dynamic Setup
          </Button>
          {fetchStudents.isError && (
            <p className="text-sm text-destructive mt-2 text-center">Failed to load course details. Please try again.</p>
          )}
        </div>
      )}

      {step === 2 && (
        <DynamicAttendance
          course={courses?.find((c: any) => c.raw_value === selectedCourse)}
          date={date}
          topic={topic}
          type={lectureType}
          extraLecture={extraLecture}
          lecNo={lecNo}
          students={students}
          onBack={() => setStep(1)}
          onSuccess={() => {
            setSuccess(true);
            queryClient.invalidateQueries({ queryKey: ["attendance-sheets"] });
            queryClient.invalidateQueries({ queryKey: ["attendance-courses"] });
            setTimeout(() => {
              setStep(1);
              setSuccess(false);
              setStudents([]);
              setSelectedCourse("");
            }, 3000);
          }}
        />
      )}
    </div>
  );
}
