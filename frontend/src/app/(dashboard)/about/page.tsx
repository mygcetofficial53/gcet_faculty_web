"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { GraduationCap, Server, Code, Shield, CheckCircle2, Heart, User, Sparkles } from "lucide-react";
import { motion } from "framer-motion";
import Link from "next/link";

const developers = [
  { name: "Yusuf Gundarwala", department: "Computer Department" },
  { name: "Abdullah Kapadia", department: "Information Technology" },
];

const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.1 },
  },
};

const item = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0 },
};

export default function AboutPage() {
  const version = "2.0.0 (Web Edition)";

  return (
    <motion.div
      variants={container}
      initial="hidden"
      animate="show"
      className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-3xl mx-auto"
    >
      {/* Hero */}
      <motion.div variants={item} className="text-center space-y-4 py-8">
        <div className="mx-auto bg-primary/10 w-24 h-24 rounded-3xl flex items-center justify-center mb-4 transform rotate-3">
          <GraduationCap className="text-primary w-14 h-14 -rotate-3" />
        </div>
        <h2 className="text-4xl font-bold font-lora text-secondary tracking-tight">GCET Faculty Portal</h2>
        <p className="text-muted-foreground max-w-lg mx-auto">
          The official digital workspace for faculty members of G H Patel College of Engineering & Technology.
        </p>
        <Badge variant="outline" className="text-primary border-primary/20 bg-primary/5 px-4 py-1">
          Version {version}
        </Badge>
      </motion.div>

      {/* Designed by */}
      <motion.div variants={item} className="text-center">
        <p className="text-sm text-muted-foreground italic">
          Designed and developed with passion by students of GCET.
        </p>
      </motion.div>

      {/* Creators Section */}
      <motion.div variants={item} className="space-y-4">
        <p className="text-xs font-bold text-primary tracking-widest uppercase ml-1">The Creators</p>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {developers.map((dev, idx) => (
            <Card key={dev.name} className="border-border/50 shadow-sm hover:shadow-md transition-shadow">
              <CardContent className="p-5 flex items-center gap-4">
                <div className="h-14 w-14 bg-primary/10 rounded-2xl flex items-center justify-center shrink-0">
                  <User className="h-7 w-7 text-primary" />
                </div>
                <div>
                  <p className="font-bold text-secondary font-lexend">{dev.name}</p>
                  <p className="text-sm text-muted-foreground">{dev.department}</p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </motion.div>

      {/* Architecture Cards */}
      <motion.div variants={item} className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Card className="border-none shadow-sm bg-blue-50/50 dark:bg-blue-900/10">
          <CardContent className="p-6 flex flex-col items-center text-center space-y-3">
            <Server className="h-8 w-8 text-blue-500" />
            <h3 className="font-semibold text-secondary">Robust Architecture</h3>
            <p className="text-sm text-muted-foreground">Powered by Go & Next.js for maximum performance and reliability.</p>
          </CardContent>
        </Card>

        <Card className="border-none shadow-sm bg-emerald-50/50 dark:bg-emerald-900/10">
          <CardContent className="p-6 flex flex-col items-center text-center space-y-3">
            <Shield className="h-8 w-8 text-emerald-500" />
            <h3 className="font-semibold text-secondary">Secure Proxy</h3>
            <p className="text-sm text-muted-foreground">Stateful session management bridging safely to the internal GMS network.</p>
          </CardContent>
        </Card>
      </motion.div>

      {/* Core Features */}
      <motion.div variants={item}>
        <Card className="shadow-sm border-border/50">
          <CardHeader>
            <CardTitle className="text-xl">Core Features</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <FeatureItem text="Automated GMS portal authentication and session management" />
            <FeatureItem text="Streamlined attendance marking with instant student fetching" />
            <FeatureItem text="Dynamic real-time attendance with code-based verification" />
            <FeatureItem text="Interactive timetable viewer with day-by-day filtering" />
            <FeatureItem text="Comprehensive analytics with subject-wise and student-wise reports" />
            <FeatureItem text="Profile management syncing with the central college database" />
          </CardContent>
        </Card>
      </motion.div>

      {/* Support CTA */}
      <motion.div variants={item}>
        <Card className="overflow-hidden border-none shadow-xl">
          <div className="bg-gradient-to-br from-primary via-orange-500 to-orange-600 p-8 text-center text-white relative overflow-hidden">
            <div className="absolute top-4 left-8 opacity-20">
              <Sparkles className="h-8 w-8" />
            </div>
            <Heart className="h-10 w-10 mx-auto mb-4 drop-shadow-md" />
            <h3 className="text-2xl font-bold font-lexend mb-2 text-white">Love the App?</h3>
            <p className="text-white/90 max-w-md mx-auto text-sm leading-relaxed mb-6">
              Support our work and help us keep the servers running.
            </p>
            <Link href="/support-developer">
              <Button
                className="bg-white text-primary hover:bg-white/90 font-bold px-8 py-3 h-auto rounded-full shadow-md"
              >
                Support Developers
              </Button>
            </Link>
          </div>
        </Card>
      </motion.div>

      {/* Footer */}
      <motion.div variants={item} className="text-center text-sm text-muted-foreground pt-4 pb-6">
        <p>&copy; {new Date().getFullYear()} G H Patel College of Engineering & Technology.</p>
        <p className="mt-1">Designed for academic excellence.</p>
      </motion.div>
    </motion.div>
  );
}

function FeatureItem({ text }: { text: string }) {
  return (
    <div className="flex items-start gap-3">
      <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0 mt-0.5" />
      <span className="text-secondary/80">{text}</span>
    </div>
  );
}
