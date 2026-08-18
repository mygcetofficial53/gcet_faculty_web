"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from "recharts";
import { useState } from "react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Loader2 } from "lucide-react";

interface SubjectAvg {
  course_code: string;
  course_name: string;
  avg_percentage: number;
}

export default function AnalyticsPage() {
  const [selectedCourse, setSelectedCourse] = useState<string>("overall");

  const { data: averages, isLoading: loadingAvg, error: errorAvg } = useQuery({
    queryKey: ["attendance-avg"],
    queryFn: async () => {
      const res = await api.get("/attendance/avg");
      if (!res.data.success) throw new Error(res.data.error || "Failed to load averages");
      return res.data.data as SubjectAvg[];
    },
  });

  const { data: courses } = useQuery({
    queryKey: ["attendance-courses"],
    queryFn: async () => {
      const res = await api.get("/attendance/courses");
      return res.data.data;
    },
  });

  const { data: studentWise, isLoading: loadingStudents } = useQuery({
    queryKey: ["student-wise", selectedCourse],
    queryFn: async () => {
      if (selectedCourse === "overall") return null;
      const course = courses?.find((c: any) => c.raw_value === selectedCourse);
      if (!course) return null;
      const res = await api.post("/attendance/student-wise", course);
      return res.data.data;
    },
    enabled: selectedCourse !== "overall" && !!courses,
  });

  if (errorAvg) {
    return (
      <div className="p-4 rounded-lg bg-destructive/10 text-destructive text-center">
        Error loading analytics: {(errorAvg as Error).message}
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h2 className="text-3xl font-bold font-lora text-secondary tracking-tight">Analytics</h2>
          <p className="text-muted-foreground mt-1">Attendance insights and reports</p>
        </div>

        <div className="w-full sm:w-72">
          <Select value={selectedCourse} onValueChange={(val) => setSelectedCourse(val || "overall")}>
            <SelectTrigger className="bg-card h-auto min-h-10 py-2 *:data-[slot=select-value]:line-clamp-none whitespace-normal text-left">
              <SelectValue placeholder="Select report type" />
            </SelectTrigger>
            <SelectContent className="max-w-[95vw] sm:max-w-3xl max-h-[60vh] !w-max min-w-[var(--anchor-width)]">
              <SelectItem value="overall" className="whitespace-normal py-2 text-sm">Overall Subject Averages</SelectItem>
              {courses?.map((c: any, idx: number) => (
                <SelectItem key={idx} value={c.raw_value} className="whitespace-normal py-2 text-sm">{c.display_text}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {selectedCourse === "overall" ? (
        <Card className="shadow-sm border-border/50">
          <CardHeader>
            <CardTitle>Subject-wise Average Attendance</CardTitle>
            <CardDescription>Overall class performance across all your assigned subjects</CardDescription>
          </CardHeader>
          <CardContent>
            {loadingAvg ? (
              <Skeleton className="h-[400px] w-full rounded-xl" />
            ) : averages && averages.length > 0 ? (
              <div className="h-[400px] w-full mt-4">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={averages} margin={{ top: 20, right: 30, left: 0, bottom: 60 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--border))" />
                    <XAxis 
                      dataKey="course_code" 
                      angle={-45} 
                      textAnchor="end"
                      height={70}
                      tick={{ fill: 'hsl(var(--muted-foreground))', fontSize: 12 }}
                      tickLine={false}
                      axisLine={false}
                    />
                    <YAxis 
                      tick={{ fill: 'hsl(var(--muted-foreground))', fontSize: 12 }}
                      tickLine={false}
                      axisLine={false}
                      tickFormatter={(value) => `${value}%`}
                    />
                    <Tooltip 
                      cursor={{ fill: 'hsl(var(--secondary)/0.05)' }}
                      contentStyle={{ borderRadius: '8px', border: '1px solid hsl(var(--border))', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                      formatter={(value: any) => [`${Number(value).toFixed(2)}%`, 'Average']}
                      labelFormatter={(label) => {
                        const subj = averages.find(a => a.course_code === label);
                        return subj ? subj.course_name : label;
                      }}
                    />
                    <Bar dataKey="avg_percentage" radius={[4, 4, 0, 0]}>
                      {averages.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.avg_percentage < 75 ? "hsl(var(--destructive))" : "hsl(var(--primary))"} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            ) : (
              <div className="text-center py-12 text-muted-foreground bg-secondary/5 rounded-xl border border-dashed border-border">
                No average attendance data available.
              </div>
            )}
          </CardContent>
        </Card>
      ) : (
        <Card className="shadow-sm border-border/50">
          <CardHeader>
            <CardTitle>Student-wise Defaulters List</CardTitle>
            <CardDescription>Students with attendance below 75% in this course</CardDescription>
          </CardHeader>
          <CardContent>
            {loadingStudents ? (
              <div className="flex justify-center p-8"><Loader2 className="h-8 w-8 animate-spin text-primary" /></div>
            ) : studentWise && studentWise.length > 0 ? (
              <div className="border rounded-xl overflow-hidden bg-card mt-4">
                <div className="max-h-[500px] overflow-y-auto custom-scrollbar">
                  <table className="w-full text-sm text-left">
                    <thead className="text-xs uppercase bg-muted/50 sticky top-0 z-10">
                      <tr>
                        <th className="px-4 py-3">Enrollment</th>
                        <th className="px-4 py-3">Student Name</th>
                        <th className="px-4 py-3 text-center">Attended/Total</th>
                        <th className="px-4 py-3 text-right">Percentage</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y">
                      {studentWise
                        .filter((s: any) => s.percentage < 75)
                        .sort((a: any, b: any) => a.percentage - b.percentage)
                        .map((student: any, idx: number) => (
                        <tr key={idx} className="hover:bg-muted/30">
                          <td className="px-4 py-3 font-medium">{student.enrollment}</td>
                          <td className="px-4 py-3">{student.name}</td>
                          <td className="px-4 py-3 text-center text-muted-foreground">
                            {student.attended} / {student.total}
                          </td>
                          <td className="px-4 py-3 text-right">
                            <span className="px-2.5 py-1 rounded-full bg-destructive/10 text-destructive font-bold">
                              {student.percentage.toFixed(2)}%
                            </span>
                          </td>
                        </tr>
                      ))}
                      {studentWise.filter((s: any) => s.percentage < 75).length === 0 && (
                        <tr>
                          <td colSpan={4} className="px-4 py-8 text-center text-muted-foreground">
                            No defaulters found! All students have attendance ≥ 75%.
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            ) : (
              <div className="text-center py-12 text-muted-foreground bg-secondary/5 rounded-xl border border-dashed border-border">
                No student data available for this course.
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
