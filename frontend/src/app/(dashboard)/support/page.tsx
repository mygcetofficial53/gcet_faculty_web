"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Loader2, Send, CheckCircle2, HelpCircle } from "lucide-react";
import { useState } from "react";

const feedbackSchema = z.object({
  type: z.string().min(1, "Please select a type"),
  subject: z.string().min(3, "Subject must be at least 3 characters"),
  description: z.string().min(10, "Description must be at least 10 characters"),
});

type FeedbackForm = z.infer<typeof feedbackSchema>;

export default function SupportPage() {
  const [success, setSuccess] = useState(false);

  const form = useForm<FeedbackForm>({
    resolver: zodResolver(feedbackSchema),
    defaultValues: {
      type: "",
      subject: "",
      description: "",
    },
  });

  const submitFeedback = useMutation({
    mutationFn: async (data: FeedbackForm) => {
      const res = await api.post("/feedback", data);
      return res.data;
    },
    onSuccess: () => {
      setSuccess(true);
      form.reset();
      setTimeout(() => setSuccess(false), 3000);
    }
  });

  const onSubmit = (data: FeedbackForm) => {
    submitFeedback.mutate(data);
  };

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-2xl mx-auto">
      <div>
        <h2 className="text-3xl font-bold font-lora text-secondary tracking-tight">Support & Feedback</h2>
        <p className="text-muted-foreground mt-1">Report issues, suggest features, or get help with the portal</p>
      </div>

      <Card className="shadow-sm border-border/50 bg-white/50 backdrop-blur-sm relative overflow-hidden">
        {/* Background icon decoration */}
        <HelpCircle className="absolute -right-8 -top-8 h-48 w-48 text-secondary/5 -z-10" />

        {success ? (
          <div className="flex flex-col items-center justify-center py-16 text-center animate-in zoom-in duration-300">
            <div className="h-20 w-20 bg-emerald-100 rounded-full flex items-center justify-center mb-6">
              <CheckCircle2 className="h-10 w-10 text-emerald-600" />
            </div>
            <h3 className="text-2xl font-bold text-secondary mb-2">Message Sent!</h3>
            <p className="text-muted-foreground">Thank you for your feedback. We will review it shortly.</p>
          </div>
        ) : (
          <>
            <CardHeader>
              <CardTitle>Submit a Ticket</CardTitle>
              <CardDescription>Fill out the form below to contact support.</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
                <div className="space-y-2">
                  <Label>Type of Request</Label>
                  <Select 
                    onValueChange={(val) => form.setValue("type", val || "")} 
                    defaultValue={form.getValues("type")}
                  >
                    <SelectTrigger className="bg-background">
                      <SelectValue placeholder="Select a category" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Bug Report">Bug Report</SelectItem>
                      <SelectItem value="Feature Request">Feature Request</SelectItem>
                      <SelectItem value="Data Correction">Data Correction</SelectItem>
                      <SelectItem value="General Query">General Query</SelectItem>
                    </SelectContent>
                  </Select>
                  {form.formState.errors.type && (
                    <p className="text-sm text-destructive">{form.formState.errors.type.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label>Subject</Label>
                  <Input 
                    placeholder="Brief summary of your request" 
                    {...form.register("subject")}
                    className="bg-background"
                  />
                  {form.formState.errors.subject && (
                    <p className="text-sm text-destructive">{form.formState.errors.subject.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label>Description</Label>
                  <Textarea 
                    placeholder="Please provide details..." 
                    className="min-h-[120px] bg-background"
                    {...form.register("description")}
                  />
                  {form.formState.errors.description && (
                    <p className="text-sm text-destructive">{form.formState.errors.description.message}</p>
                  )}
                </div>
              </form>
            </CardContent>
            <CardFooter>
              <Button 
                onClick={form.handleSubmit(onSubmit)} 
                disabled={submitFeedback.isPending}
                className="w-full sm:w-auto"
              >
                {submitFeedback.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Send className="mr-2 h-4 w-4" />}
                Submit Ticket
              </Button>
            </CardFooter>
          </>
        )}
      </Card>
    </div>
  );
}
