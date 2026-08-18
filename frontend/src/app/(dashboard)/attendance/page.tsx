"use client";


import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card } from "@/components/ui/card";
import { ClipboardCheck, FileEdit, Trash2, FileSpreadsheet, Radio } from "lucide-react";
import ViewAttendanceSheets from "./components/view-sheets";
import EnterAttendance from "./components/enter-attendance";
import DynamicSessionTab from "./components/dynamic-session-tab";

export default function AttendancePage() {
  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div>
        <h2 className="text-3xl font-bold font-lora text-secondary tracking-tight">Manage Attendance</h2>
        <p className="text-muted-foreground mt-1">View sheets, enter, edit, or delete attendance records</p>
      </div>

      <Card className="shadow-sm border-border/50">
        <Tabs defaultValue="enter" className="w-full">
          <div className="border-b px-2 pt-2">
            <TabsList className="w-full justify-start h-auto bg-transparent overflow-x-auto flex-nowrap hide-scrollbar">
              <TabsTrigger 
                value="enter"
                className="flex items-center gap-2 px-4 py-2.5 data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent data-[state=active]:bg-transparent"
              >
                <ClipboardCheck className="h-4 w-4" />
                Manual Entry
              </TabsTrigger>
              <TabsTrigger 
                value="dynamic"
                className="flex items-center gap-2 px-4 py-2.5 data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent data-[state=active]:bg-transparent text-emerald-600 data-[state=active]:text-emerald-600"
              >
                <Radio className="h-4 w-4" />
                Dynamic Session
              </TabsTrigger>
              <TabsTrigger 
                value="view"
                className="flex items-center gap-2 px-4 py-2.5 data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent data-[state=active]:bg-transparent"
              >
                <FileSpreadsheet className="h-4 w-4" />
                View Sheets
              </TabsTrigger>
              <TabsTrigger 
                value="edit"
                className="flex items-center gap-2 px-4 py-2.5 data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent data-[state=active]:bg-transparent"
              >
                <FileEdit className="h-4 w-4" />
                Edit
              </TabsTrigger>
              <TabsTrigger 
                value="delete"
                className="flex items-center gap-2 px-4 py-2.5 data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent data-[state=active]:bg-transparent"
              >
                <Trash2 className="h-4 w-4" />
                Delete
              </TabsTrigger>
            </TabsList>
          </div>

          <div className="p-6">
            <TabsContent value="enter" className="m-0 focus-visible:outline-none">
              <EnterAttendance />
            </TabsContent>
            
            <TabsContent value="dynamic" className="m-0 focus-visible:outline-none">
              <DynamicSessionTab />
            </TabsContent>
            
            <TabsContent value="view" className="m-0 focus-visible:outline-none">
              <ViewAttendanceSheets />
            </TabsContent>

            <TabsContent value="edit" className="m-0 focus-visible:outline-none">
              <div className="text-center py-12 text-muted-foreground bg-secondary/5 rounded-xl border border-dashed border-border">
                Edit attendance feature coming soon.
              </div>
            </TabsContent>

            <TabsContent value="delete" className="m-0 focus-visible:outline-none">
              <div className="text-center py-12 text-muted-foreground bg-secondary/5 rounded-xl border border-dashed border-border">
                Delete attendance feature coming soon.
              </div>
            </TabsContent>
          </div>
        </Tabs>
      </Card>
    </div>
  );
}
