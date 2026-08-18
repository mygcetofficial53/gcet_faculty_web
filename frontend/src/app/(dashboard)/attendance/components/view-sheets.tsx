"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { BookOpen } from "lucide-react";

interface AttendanceSheet {
  course_code: string;
  course_name: string;
  type: string;
  batch: string;
  average_attendance: number;
}

export default function ViewAttendanceSheets() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["attendance-sheets"],
    queryFn: async () => {
      const res = await api.get("/attendance/sheets");
      if (!res.data.success) throw new Error(res.data.error || "Failed to load sheets");
      return res.data.data as AttendanceSheet[];
    },
  });

  if (error) {
    return (
      <div className="p-4 rounded-lg bg-destructive/10 text-destructive text-center">
        Error loading sheets: {(error as Error).message}
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3, 4].map((i) => (
          <Skeleton key={i} className="h-16 w-full rounded-lg" />
        ))}
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground bg-secondary/5 rounded-xl border border-dashed border-border">
        No attendance sheets found for your courses.
      </div>
    );
  }

  return (
    <div className="rounded-xl border border-border/50 overflow-hidden">
      <Table>
        <TableHeader className="bg-secondary/5">
          <TableRow>
            <TableHead>Course</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Batch</TableHead>
            <TableHead className="text-right">Avg Attendance</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((sheet, idx) => (
            <TableRow key={idx} className="hover:bg-secondary/5">
              <TableCell>
                <div className="flex items-start gap-3">
                  <div className="p-2 bg-primary/10 rounded-lg mt-0.5">
                    <BookOpen className="h-4 w-4 text-primary" />
                  </div>
                  <div>
                    <p className="font-medium text-secondary">{sheet.course_name}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">{sheet.course_code}</p>
                  </div>
                </div>
              </TableCell>
              <TableCell>
                <span className="px-2.5 py-1 rounded-full bg-secondary/10 text-secondary text-xs font-medium">
                  {sheet.type}
                </span>
              </TableCell>
              <TableCell>{sheet.batch || "All"}</TableCell>
              <TableCell className="text-right font-medium">
                <span className={sheet.average_attendance < 75 ? "text-destructive" : "text-emerald-600"}>
                  {sheet.average_attendance.toFixed(2)}%
                </span>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
