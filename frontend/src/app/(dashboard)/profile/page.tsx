"use client";

import { useAuthStore } from "@/store/useAuthStore";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Mail, Phone, MapPin, CalendarDays, Briefcase, GraduationCap, Building2 } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function ProfilePage() {
  const user = useAuthStore((state) => state.user);

  if (!user) return null;

  const getInitials = (name?: string) => {
    if (!name) return "F";
    const parts = name.split(" ");
    if (parts.length >= 2) return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
    return name.substring(0, 2).toUpperCase();
  };

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-4xl mx-auto">
      <Card className="border-none shadow-md overflow-hidden bg-white/50 backdrop-blur-sm">
        <div className="h-32 bg-gradient-to-r from-primary/80 to-primary w-full relative">
          <div className="absolute inset-0 bg-academic-grid opacity-20"></div>
        </div>
        <CardContent className="px-6 pb-6 relative">
          <div className="flex flex-col sm:flex-row items-center sm:items-end gap-6 -mt-16 sm:-mt-12 mb-6">
            <Avatar className="h-32 w-32 border-4 border-background shadow-lg">
              <AvatarImage src={`https://api.dicebear.com/7.x/initials/svg?seed=${user.name}&backgroundColor=D35D27&textColor=ffffff`} />
              <AvatarFallback className="text-4xl">{getInitials(user.name)}</AvatarFallback>
            </Avatar>
            <div className="text-center sm:text-left space-y-1 mb-2">
              <h2 className="text-2xl font-bold font-lora text-secondary">{user.name}</h2>
              <div className="flex flex-wrap items-center justify-center sm:justify-start gap-2 text-muted-foreground font-medium">
                <span className="flex items-center gap-1.5"><Briefcase className="h-4 w-4" /> {user.designation || "Faculty"}</span>
                <span className="hidden sm:inline">•</span>
                <span className="flex items-center gap-1.5"><Building2 className="h-4 w-4" /> {user.department}</span>
              </div>
            </div>
            <div className="sm:ml-auto flex gap-2">
              <Badge variant="secondary" className="px-3 py-1 bg-primary/10 text-primary hover:bg-primary/20">
                Emp ID: {user.employee_id}
              </Badge>
            </div>
          </div>

          <Tabs defaultValue="personal" className="w-full">
            <TabsList className="w-full justify-start h-auto p-1 bg-secondary/5 overflow-x-auto flex-nowrap hide-scrollbar mb-6">
              <TabsTrigger value="personal" className="px-6">Personal Info</TabsTrigger>
              <TabsTrigger value="academic" className="px-6">Academic Info</TabsTrigger>
            </TabsList>

            <TabsContent value="personal" className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <h3 className="font-semibold text-secondary flex items-center gap-2 border-b pb-2">
                    <UserIcon className="h-4 w-4 text-primary" /> Contact Details
                  </h3>
                  <div className="space-y-3">
                    <InfoRow icon={<Mail className="h-4 w-4 text-muted-foreground" />} label="Email" value={user.email} />
                    <InfoRow icon={<Phone className="h-4 w-4 text-muted-foreground" />} label="Phone" value={user.phone} />
                    <InfoRow icon={<MapPin className="h-4 w-4 text-muted-foreground" />} label="Address" value={user.address} />
                  </div>
                </div>
                
                <div className="space-y-4">
                  <h3 className="font-semibold text-secondary flex items-center gap-2 border-b pb-2">
                    <CalendarDays className="h-4 w-4 text-primary" /> Identity
                  </h3>
                  <div className="space-y-3">
                    <InfoRow label="Date of Birth" value={user.dob} />
                    <InfoRow label="Gender" value={user.gender} />
                    <InfoRow label="Blood Group" value={user.blood_group} />
                    <InfoRow label="Marital Status" value={user.marital_status} />
                  </div>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="academic" className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <h3 className="font-semibold text-secondary flex items-center gap-2 border-b pb-2">
                    <GraduationCap className="h-4 w-4 text-primary" /> Qualifications
                  </h3>
                  <div className="space-y-3">
                    <InfoRow label="Highest Qualification" value={user.qualification || "Ph.D / Masters"} />
                    <InfoRow label="Experience" value={user.experience ? `${user.experience} Years` : "N/A"} />
                    <InfoRow label="Joining Date" value={user.joining_date} />
                  </div>
                </div>

                <div className="space-y-4">
                  <h3 className="font-semibold text-secondary flex items-center gap-2 border-b pb-2">
                    <BookOpenIcon className="h-4 w-4 text-primary" /> Subjects Taught
                  </h3>
                  {user.subjects && user.subjects.length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                      {user.subjects.map((sub, i) => (
                        <Badge key={i} variant="outline" className="bg-background text-xs font-normal py-1 border-border/60">
                          {sub}
                        </Badge>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground italic">No subjects listed currently.</p>
                  )}
                </div>
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}

function InfoRow({ icon, label, value }: { icon?: React.ReactNode; label: string; value?: string }) {
  if (!value) return null;
  return (
    <div className="flex items-start gap-3 text-sm">
      {icon && <div className="mt-0.5">{icon}</div>}
      <div>
        <p className="text-muted-foreground text-xs font-medium">{label}</p>
        <p className="text-secondary font-medium">{value}</p>
      </div>
    </div>
  );
}

// Icons not in the main lucide import
function UserIcon(props: any) {
  return <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
}
function BookOpenIcon(props: any) {
  return <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
}
