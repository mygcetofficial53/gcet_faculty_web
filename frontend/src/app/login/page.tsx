"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { api } from "@/lib/api";
import Cookies from 'js-cookie';
import { useAuthStore } from "@/store/useAuthStore";
import { useRouter } from "next/navigation";
import { FloatingBackground } from "@/components/ui/floating-background";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Loader2 } from "lucide-react";
import { motion } from "framer-motion";
import Image from "next/image";

const loginSchema = z.object({
  login_id: z.string().min(1, "Login ID is required"),
  password: z.string().min(1, "Password is required"),
});

type LoginForm = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const setAuth = useAuthStore((state) => state.setAuth);
  const [error, setError] = useState<string | null>(null);

  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      login_id: "",
      password: "",
    },
  });

  const onSubmit = async (data: LoginForm) => {
    setError(null);
    try {
      const res = await api.post("/auth/login", data);
      if (res.data.success) {
        // Store JWT tokens in cookies for authenticated API calls
        if (res.data.token) {
          Cookies.set('token', res.data.token, { secure: true, sameSite: 'strict' });
        }
        if (res.data.refresh_token) {
          Cookies.set('refresh_token', res.data.refresh_token, { secure: true, sameSite: 'strict' });
        }
        setAuth(res.data.faculty);
        router.push("/");
      } else {
        setError(res.data.error || "Login failed");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Unable to connect to GMS Portal. Please try again.");
    }
  };

  return (
    <main className="min-h-screen w-full flex items-center justify-center relative px-4">
      <FloatingBackground />
      
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: "easeOut" }}
        className="w-full max-w-md z-10"
      >
        <Card className="border-none shadow-2xl bg-white/95 backdrop-blur-md overflow-hidden">
          {/* Top brand accent */}
          <div className="h-2 w-full bg-primary" />
          
          <CardHeader className="space-y-3 pb-6 pt-8 text-center">
            <motion.div
              initial={{ scale: 0.8 }}
              animate={{ scale: 1 }}
              transition={{ delay: 0.2, type: "spring" }}
              className="mx-auto mb-2"
            >
              <div className="w-[100px] h-[100px] rounded-full overflow-hidden border-2 border-primary/20 bg-white shadow-md mx-auto">
                <Image 
                  src="/logo.png" 
                  alt="GCET Logo" 
                  width={100} 
                  height={100} 
                  className="object-cover scale-110"
                />
              </div>
            </motion.div>
            <CardTitle className="text-3xl font-bold font-lora text-secondary">
              Faculty Dashboard.
            </CardTitle>
            <p className="text-sm font-medium text-primary/80 font-inter leading-relaxed">
              G.H. Patel College of Engineering<br/>and Technology.
            </p>
            <CardDescription className="text-base text-muted-foreground font-inter pt-2">
              Sign in with your GMS credentials
            </CardDescription>
          </CardHeader>
          
          <CardContent>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="login_id">Login ID</Label>
                <Input
                  id="login_id"
                  autoComplete="username"
                  placeholder="e.g. j.doe"
                  className="h-12 bg-background/50"
                  {...form.register("login_id")}
                />
                {form.formState.errors.login_id && (
                  <p className="text-sm text-destructive">{form.formState.errors.login_id.message}</p>
                )}
              </div>
              
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  autoComplete="current-password"
                  placeholder="••••••••"
                  className="h-12 bg-background/50"
                  {...form.register("password")}
                />
                {form.formState.errors.password && (
                  <p className="text-sm text-destructive">{form.formState.errors.password.message}</p>
                )}
              </div>

              {error && (
                <motion.div 
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: "auto" }}
                  className="p-3 rounded-lg bg-destructive/10 text-destructive text-sm"
                >
                  {error}
                </motion.div>
              )}

              <Button 
                type="submit" 
                className="w-full h-12 text-lg font-medium shadow-lg hover:shadow-primary/25 transition-all mt-4"
                disabled={form.formState.isSubmitting}
              >
                {form.formState.isSubmitting ? (
                  <>
                    <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                    Connecting to GMS...
                  </>
                ) : (
                  "Sign In"
                )}
              </Button>
            </form>
          </CardContent>
          <CardFooter className="justify-center pb-8 pt-2">
            <p className="text-sm text-muted-foreground">
              Securely proxies to GCET GMS Portal
            </p>
          </CardFooter>
        </Card>
      </motion.div>
    </main>
  );
}
