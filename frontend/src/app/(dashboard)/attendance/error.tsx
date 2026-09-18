"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { AlertCircle } from "lucide-react";

export default function AttendanceError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Attendance Page Error:", error);
  }, [error]);

  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] p-8 text-center bg-destructive/5 rounded-xl border border-destructive/20 animate-in fade-in">
      <div className="p-4 bg-destructive/10 rounded-full mb-4">
        <AlertCircle className="h-8 w-8 text-destructive" />
      </div>
      <h2 className="text-xl font-bold text-destructive mb-2">Something went wrong!</h2>
      <p className="text-muted-foreground mb-4">
        The attendance module failed to load.
      </p>
      
      {/* Show the exact error message */}
      <div className="bg-background border rounded p-4 mb-6 text-left max-w-2xl overflow-auto text-sm font-mono text-destructive">
        <strong>Error:</strong> {error.message}
        {error.stack && (
          <pre className="mt-2 text-xs opacity-80 whitespace-pre-wrap">
            {error.stack}
          </pre>
        )}
      </div>

      <Button variant="default" onClick={() => reset()}>
        Try again
      </Button>
    </div>
  );
}
