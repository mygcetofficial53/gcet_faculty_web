"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Slider } from "@/components/ui/slider";
import { supabase } from "@/lib/supabase";
import { api } from "@/lib/api";
import { useToast } from "@/components/ui/toast";
import { Loader2, Radio, CheckCircle2, Users, BadgePlus, XCircle, Info, Keyboard } from "lucide-react";

interface DynamicAttendanceProps {
  course: any;
  date: string;
  topic: string;
  type: string;
  extraLecture: boolean;
  lecNo: string;
  students: any[]; // { sr_no, enrollment, name, is_present }
  onBack: () => void;
  onSuccess: () => void;
}

export function DynamicAttendance({ course, date, topic, type, extraLecture, lecNo, students, onBack, onSuccess }: DynamicAttendanceProps) {
  const toast = useToast();
  const [isSixDigits, setIsSixDigits] = useState(false);
  const [durationSeconds, setDurationSeconds] = useState(30);
  
  const [isGenerating, setIsGenerating] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [generatedCode, setGeneratedCode] = useState<string | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [secondsRemaining, setSecondsRemaining] = useState(0);
  
  const [manualRoll, setManualRoll] = useState("");
  const [isAddingManual, setIsAddingManual] = useState(false);
  const [presentStudents, setPresentStudents] = useState<any[]>([]); // list of student objects

  // Generate a random numeric code
  const generateCode = (length: number) => {
    let code = '';
    for (let i = 0; i < length; i++) {
      code += Math.floor(Math.random() * 10).toString();
    }
    return code;
  };

  const startSession = async () => {
    if (!students || students.length === 0) {
      toast.warning("No Students", "No students loaded for this course!");
      return;
    }

    setIsGenerating(true);
    try {
      // 1. Sync Enrollments
      const enrollments = students.map(s => ({
        course_id: course.raw_value || course.course_code,
        student_roll_no: s.enrollment
      }));
      const { error: upsertErr } = await supabase.from('enrollments').upsert(enrollments, { onConflict: 'course_id, student_roll_no' });
      if (upsertErr) throw upsertErr;

      // 2. Start Session via RPC
      const code = generateCode(isSixDigits ? 6 : 4);
      
      const { data: sessionData, error } = await supabase.rpc('create_attendance_session', {
        p_course_id: course.raw_value || course.course_code,
        p_course_name: course.display_text || course.courseName,
        p_batch: course.metadata?.batch || course.batch,
        p_semester: course.metadata?.semester,
        p_code: code,
        p_duration_seconds: durationSeconds,
        p_session_type: type
      });
      
      if (error) throw error;
      
      if (!sessionData) throw new Error("No session data returned");

      setSessionId(sessionData.id);
      setGeneratedCode(sessionData.dynamic_code || code);
      setSecondsRemaining(durationSeconds);

    } catch (e: any) {
      console.error(e);
      toast.error("Session Failed", e.message || "Failed to start session.");
    } finally {
      setIsGenerating(false);
    }
  };

  // Timer effect
  useEffect(() => {
    let interval: any;
    if (secondsRemaining > 0 && generatedCode) {
      interval = setInterval(() => {
        setSecondsRemaining(prev => {
          if (prev <= 1) {
            clearInterval(interval);
            endSessionAndSubmit();
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    }
    return () => clearInterval(interval);
  }, [secondsRemaining, generatedCode]);

  // Realtime subscription effect
  useEffect(() => {
    if (!sessionId) return;

    const channel = supabase
      .channel('attendance_changes')
      .on('postgres_changes', 
        { event: 'INSERT', schema: 'public', table: 'attendance_records', filter: `session_id=eq.${sessionId}` },
        (payload) => {
          const newRecord = payload.new;
          // Find student in our original list
          const studentInfo = students.find(s => s.enrollment === newRecord.student_roll_no);
          if (studentInfo) {
            setPresentStudents(prev => {
              if (prev.find(p => p.enrollment === studentInfo.enrollment)) return prev; // Avoid duplicates
              return [studentInfo, ...prev]; // Add to top
            });
          }
        }
      )
      .subscribe();

    return () => {
      supabase.removeChannel(channel);
    };
  }, [sessionId, students]);

  const endSessionAndSubmit = async () => {
    if (isSubmitting) return;
    setIsSubmitting(true);
    try {
      // Map back to the students array format required by backend
      const finalStudents = students.map(s => ({
        ...s,
        is_present: !!presentStudents.find(p => p.enrollment === s.enrollment)
      }));

      const res = await api.post("/attendance/enter", {
        course_code: course.course_code,
        date: date,
        time_slot: "", 
        topic: topic,
        type: type,
        extra_lecture: extraLecture,
        lec_no: lecNo,
        metadata: course.metadata,
        option_index: course.option_index,
        students: finalStudents
      });

      if (!res.data.success) {
        throw new Error(res.data.error || "Failed to submit attendance to server");
      }
      onSuccess();
    } catch (e: any) {
      console.error(e);
      toast.error("Submission Failed", e.message || "Failed to submit final attendance.");
      setIsSubmitting(false);
    }
  };

  const addManualStudent = async () => {
    if (!manualRoll || !sessionId) return;
    setIsAddingManual(true);
    try {
      const studentInfo = students.find(s => s.enrollment === manualRoll);
      if (!studentInfo) {
        toast.warning("Not Enrolled", `Student ${manualRoll} is not enrolled in this course.`);
        setIsAddingManual(false);
        return;
      }
      // Call Supabase RPC to add attendance manually
      const { error } = await supabase.rpc('add_manual_attendance', {
        p_session_id: sessionId,
        p_student_roll_no: manualRoll,
        p_device_id: 'MANUAL_' + manualRoll,
        p_manual_reason: 'Teacher Override Web'
      });
      if (error) throw error;
      setManualRoll("");
    } catch (e: any) {
      toast.error("Manual Entry Failed", e.message || "Error adding manual student.");
    } finally {
      setIsAddingManual(false);
    }
  };

  const cancelSession = () => {
    if (window.confirm("Are you sure you want to cancel this dynamic session? No attendance will be saved.")) {
      setGeneratedCode(null);
      setSessionId(null);
      setSecondsRemaining(0);
      onBack();
    }
  };

  if (!generatedCode) {
    return (
      <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <div className="flex items-center gap-4 border-b pb-4">
          <Button variant="ghost" onClick={onBack}>← Back to Manual</Button>
          <div>
            <h2 className="text-xl font-bold text-secondary">Setup Dynamic Session</h2>
            <p className="text-sm text-muted-foreground">{course.display_text}</p>
          </div>
        </div>

        <Card className="max-w-2xl mx-auto border-border/50 shadow-sm mt-8">
          <CardContent className="p-8 space-y-8">
            <div className="space-y-8">
              <div className="space-y-4">
                <div>
                  <Label className="text-xs font-bold text-primary tracking-widest uppercase">Code Length</Label>
                </div>
                <div className="flex gap-4">
                  <div 
                    onClick={() => setIsSixDigits(false)}
                    className={`flex-1 py-4 text-center rounded-2xl border-2 cursor-pointer transition-all ${!isSixDigits ? 'border-primary bg-primary text-white shadow-md' : 'border-gray-200 text-muted-foreground'}`}
                  >
                    <span className="font-bold">4 Digits</span>
                  </div>
                  <div 
                    onClick={() => setIsSixDigits(true)}
                    className={`flex-1 py-4 text-center rounded-2xl border-2 cursor-pointer transition-all ${isSixDigits ? 'border-primary bg-primary text-white shadow-md' : 'border-gray-200 text-muted-foreground'}`}
                  >
                    <span className="font-bold">6 Digits</span>
                  </div>
                </div>
              </div>

              <div className="space-y-4 pt-6 border-t border-border/50">
                <div>
                  <Label className="text-xs font-bold text-primary tracking-widest uppercase">Session Duration</Label>
                </div>
                <div className="flex gap-4">
                  {[10, 15, 30].map(duration => (
                    <div 
                      key={duration}
                      onClick={() => setDurationSeconds(duration)}
                      className={`flex-1 py-3 text-center rounded-2xl border-2 cursor-pointer transition-all ${durationSeconds === duration ? 'border-primary bg-primary text-white shadow-md' : 'border-gray-200 text-muted-foreground'}`}
                    >
                      <span className="font-bold">{duration}s</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <Button 
              className="w-full h-14 text-lg font-bold" 
              onClick={startSession}
              disabled={isGenerating}
            >
              {isGenerating ? <Loader2 className="h-5 w-5 animate-spin mr-2" /> : <Radio className="h-5 w-5 mr-2" />}
              Start Broadcasting
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Active Session View

  return (
    <div className="space-y-8 animate-in fade-in zoom-in-95 duration-500">
      <div className="text-center bg-card p-12 rounded-3xl border shadow-sm max-w-4xl mx-auto mt-8">
        <h2 className="text-sm font-bold text-primary tracking-widest uppercase mb-8">
          {secondsRemaining === 0 ? "Session Expired" : "Dynamic Attendance Code"}
        </h2>
        <div className="text-8xl font-black tracking-[0.2em] font-mono text-secondary mb-12 drop-shadow-sm transition-all duration-300">
          {generatedCode}
        </div>
        
        {secondsRemaining > 0 && (
          <div className="flex items-center justify-center gap-2 mb-10">
            <Radio className="h-4 w-4 text-emerald-500 animate-pulse" />
            <p className="text-emerald-600 font-bold text-sm tracking-wide">Broadcasting on Realtime Channel...</p>
          </div>
        )}

        <div className="relative w-48 h-48 mx-auto flex items-center justify-center">
          <svg className="absolute inset-0 w-full h-full transform -rotate-90">
            <circle cx="96" cy="96" r="84" stroke="currentColor" strokeWidth="16" fill="transparent" className="text-secondary/10" />
            <circle cx="96" cy="96" r="84" stroke="currentColor" strokeWidth="16" fill="transparent" strokeLinecap="round"
              strokeDasharray={528}
              strokeDashoffset={528 - (528 * (secondsRemaining / durationSeconds))}
              className={`${secondsRemaining === 0 ? 'text-secondary/20' : 'text-primary'} transition-all duration-1000 ease-linear`}
            />
          </svg>
          <div className="flex flex-col items-center">
            <span className={`text-5xl font-black ${secondsRemaining === 0 ? 'text-secondary/40' : 'text-secondary'}`}>
              {secondsRemaining}
            </span>
            <span className="text-xs text-muted-foreground font-bold tracking-widest mt-1">SEC LEFT</span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-5xl mx-auto mt-12">
        <div className="space-y-8">
          {/* Manual Entry Block */}
          <Card className="border-border/50 shadow-sm overflow-hidden">
            <div className="bg-primary/5 px-6 py-4 border-b flex items-center gap-3">
              <div className="p-2 bg-primary/10 rounded-lg text-primary">
                <Keyboard className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-xs font-bold text-primary tracking-widest">MANUAL ENTRY</h3>
                <p className="font-semibold text-secondary">Add Students</p>
              </div>
            </div>
            <CardContent className="p-6 space-y-4">
              <input
                type="text"
                placeholder="Enter Enrollment No."
                value={manualRoll}
                onChange={(e) => setManualRoll(e.target.value)}
                className="w-full bg-secondary/5 border-none h-14 rounded-xl px-4 font-mono font-bold text-lg"
              />
              <Button 
                className="w-full h-12 text-base font-bold bg-secondary hover:bg-secondary/90" 
                onClick={addManualStudent}
                disabled={isAddingManual}
              >
                {isAddingManual ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <BadgePlus className="h-4 w-4 mr-2" />}
                Add Student
              </Button>
              <div className="flex items-start gap-2 bg-primary/5 p-3 rounded-lg mt-4">
                <Info className="h-4 w-4 text-primary mt-0.5" />
                <p className="text-xs text-primary font-medium">Teacher override records are cross-logged to the database.</p>
              </div>
            </CardContent>
          </Card>

          <div className="flex flex-col gap-4">
            <Button 
              variant="default" 
              size="lg" 
              className="h-16 text-lg font-bold bg-emerald-600 hover:bg-emerald-700 w-full shadow-lg shadow-emerald-600/20"
              onClick={endSessionAndSubmit}
              disabled={isSubmitting}
            >
              {isSubmitting ? (
                <><Loader2 className="h-5 w-5 animate-spin mr-2" /> Syncing to Portal...</>
              ) : (
                "Sync Attendance with Portal"
              )}
            </Button>
            <Button 
              variant="outline" 
              size="lg" 
              className="h-14 text-base w-full text-destructive border-destructive/20 hover:bg-destructive/5 hover:text-destructive"
              onClick={cancelSession}
              disabled={isSubmitting}
            >
              Cancel Session
            </Button>
          </div>
        </div>

        {/* Present Students Block */}
        <Card className="border-border/50 shadow-sm h-[500px] flex flex-col">
          <CardContent className="p-0 flex-1 flex flex-col min-h-0">
            <div className="p-6 border-b flex items-center justify-between">
              <div>
                <h3 className="text-xs font-bold text-primary tracking-widest">STUDENTS PRESENT</h3>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-3xl font-black text-secondary">{presentStudents.length}</span>
                <div className="p-3 bg-primary/10 rounded-full text-primary">
                  <Users className="h-5 w-5" />
                </div>
              </div>
            </div>
            
            <div className="flex-1 overflow-y-auto p-4 space-y-3 custom-scrollbar">
              {presentStudents.length === 0 ? (
                <div className="h-full flex flex-col items-center justify-center text-muted-foreground opacity-50">
                  <Loader2 className="h-8 w-8 animate-spin mb-4 text-primary" />
                  <p>Waiting for students to join...</p>
                </div>
              ) : (
                presentStudents.map((s, i) => (
                  <div key={i} className="flex items-center gap-4 bg-secondary/5 p-4 rounded-xl animate-in fade-in slide-in-from-left-4">
                    <CheckCircle2 className="h-6 w-6 text-emerald-500 flex-shrink-0 drop-shadow-sm" />
                    <div>
                      <p className="font-bold text-sm text-secondary line-clamp-1">{s.name}</p>
                      <p className="text-xs text-muted-foreground font-mono mt-0.5">{s.enrollment}</p>
                    </div>
                  </div>
                ))
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
