"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Bell, Moon, Shield, Smartphone } from "lucide-react";
import { useState } from "react";

export default function SettingsPage() {
  const [pushEnabled, setPushEnabled] = useState(true);
  const [darkMode, setDarkMode] = useState(false);
  
  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-3xl mx-auto">
      <div>
        <h2 className="text-3xl font-bold font-lora text-secondary tracking-tight">Settings</h2>
        <p className="text-muted-foreground mt-1">Manage your application preferences</p>
      </div>

      <div className="space-y-4">
        <Card className="shadow-sm border-border/50">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Bell className="h-5 w-5 text-primary" /> Notifications
            </CardTitle>
            <CardDescription>Configure how you want to be alerted.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label className="text-base">Push Notifications</Label>
                <p className="text-sm text-muted-foreground">Receive browser notifications for class reminders.</p>
              </div>
              <Switch checked={pushEnabled} onCheckedChange={setPushEnabled} />
            </div>
          </CardContent>
        </Card>

        <Card className="shadow-sm border-border/50">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Moon className="h-5 w-5 text-primary" /> Appearance
            </CardTitle>
            <CardDescription>Customize the look and feel of the portal.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label className="text-base">Dark Mode</Label>
                <p className="text-sm text-muted-foreground">Switch to a darker theme (Coming soon).</p>
              </div>
              <Switch checked={darkMode} onCheckedChange={setDarkMode} disabled />
            </div>
          </CardContent>
        </Card>

        <Card className="shadow-sm border-border/50 opacity-70">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Shield className="h-5 w-5 text-primary" /> Security
            </CardTitle>
            <CardDescription>Manage your security preferences.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label className="text-base">Change GMS Password</Label>
                <p className="text-sm text-muted-foreground">Must be changed through the official GMS portal directly.</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
