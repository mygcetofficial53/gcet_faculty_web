"use client";

import { useState } from "react";
import Script from "next/script";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Heart, IndianRupee, CheckCircle2, XCircle, Sparkles, Coffee, Rocket, Gem, User } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

declare global {
  interface Window {
    Razorpay: any;
  }
}

const RAZORPAY_KEY = process.env.NEXT_PUBLIC_RAZORPAY_KEY || "";

const PRESET_AMOUNTS = [
  { amount: 53, icon: Coffee, description: "Buy us a coffee" },
  { amount: 153, icon: Rocket, description: "Keep the servers running" },
  { amount: 253, icon: Gem, description: "Support premium features" },
];

const developers = [
  { name: "Yusuf Gundarwala", department: "Computer Department" },
  { name: "Abdullah Kapadia", department: "Information Technology" },
];

export default function SupportDeveloperPage() {
  const [customAmount, setCustomAmount] = useState("");
  const [paymentState, setPaymentState] = useState<"idle" | "success" | "error">("idle");
  const [razorpayReady, setRazorpayReady] = useState(false);

  const initiatePayment = (amountInRupees: number) => {
    if (!window.Razorpay || !razorpayReady) {
      return;
    }

    if (amountInRupees < 1) {
      alert("Please enter a valid amount.");
      return;
    }

    const options = {
      key: RAZORPAY_KEY,
      amount: amountInRupees * 100, // Amount in paisa
      name: "GCET Faculty App",
      description: "Support Developers",
      timeout: 300,
      theme: {
        color: "#D35D27",
      },
      handler: () => {
        setPaymentState("success");
        setTimeout(() => setPaymentState("idle"), 5000);
      },
      modal: {
        ondismiss: () => {
          // User closed the modal without paying
        },
      },
    };

    try {
      const rzp = new window.Razorpay(options);
      rzp.on("payment.failed", () => {
        setPaymentState("error");
        setTimeout(() => setPaymentState("idle"), 4000);
      });
      rzp.open();
    } catch (e) {
      console.error("Error opening Razorpay:", e);
      setPaymentState("error");
      setTimeout(() => setPaymentState("idle"), 4000);
    }
  };

  const handleCustomPay = () => {
    const amt = parseInt(customAmount.trim());
    if (isNaN(amt) || amt <= 0) {
      alert("Please enter a valid amount.");
      return;
    }
    initiatePayment(amt);
  };

  return (
    <>
      <Script src="https://checkout.razorpay.com/v1/checkout.js" strategy="afterInteractive" onLoad={() => setRazorpayReady(true)} />
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-3xl mx-auto">
      {/* Header */}
      <div className="text-center space-y-3 pt-4">
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ type: "spring", duration: 0.6 }}
          className="mx-auto w-20 h-20 bg-gradient-to-br from-primary to-orange-600 rounded-3xl flex items-center justify-center shadow-lg shadow-primary/30"
        >
          <Heart className="h-10 w-10 text-white" />
        </motion.div>
        <h2 className="text-3xl font-bold font-lora text-secondary tracking-tight">
          Support the Developers
        </h2>
        <p className="text-muted-foreground max-w-md mx-auto">
          Your contribution helps us keep the app maintained, ad-free, and constantly improving. Thank you! 🧡
        </p>
      </div>

      {/* Payment Status Banners */}
      <AnimatePresence mode="wait">
        {paymentState === "success" && (
          <motion.div
            key="success"
            initial={{ opacity: 0, y: -20, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -20, scale: 0.95 }}
            className="bg-emerald-50 border border-emerald-200 rounded-2xl p-6 flex items-center gap-4"
          >
            <div className="h-14 w-14 bg-emerald-100 rounded-full flex items-center justify-center shrink-0">
              <CheckCircle2 className="h-7 w-7 text-emerald-600" />
            </div>
            <div>
              <h3 className="font-bold text-emerald-800 text-lg">Payment Successful!</h3>
              <p className="text-emerald-600 text-sm">Thank you for your generous support. You're amazing! 🎉</p>
            </div>
          </motion.div>
        )}
        {paymentState === "error" && (
          <motion.div
            key="error"
            initial={{ opacity: 0, y: -20, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -20, scale: 0.95 }}
            className="bg-red-50 border border-red-200 rounded-2xl p-6 flex items-center gap-4"
          >
            <div className="h-14 w-14 bg-red-100 rounded-full flex items-center justify-center shrink-0">
              <XCircle className="h-7 w-7 text-red-600" />
            </div>
            <div>
              <h3 className="font-bold text-red-800 text-lg">Payment Failed</h3>
              <p className="text-red-600 text-sm">The payment was cancelled or encountered an error. Please try again.</p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Preset Amount Cards */}
      <div>
        <p className="text-xs font-bold text-primary tracking-widest uppercase mb-4 text-center">Choose an Amount</p>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          {PRESET_AMOUNTS.map((preset, idx) => (
            <motion.div
              key={preset.amount}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: idx * 0.1 }}
            >
              <Card
                className={`border-2 border-transparent transition-all duration-300 group ${
                  razorpayReady
                    ? "cursor-pointer hover:border-primary/40 hover:shadow-lg hover:shadow-primary/10"
                    : "opacity-50 cursor-not-allowed"
                }`}
                onClick={() => razorpayReady && initiatePayment(preset.amount)}
              >
                <CardContent className="p-6 text-center space-y-3">
                  <preset.icon className="h-8 w-8 mx-auto text-primary" />
                  <div>
                    <p className="text-3xl font-black text-secondary group-hover:text-primary transition-colors">
                      ₹{preset.amount}
                    </p>
                    <p className="text-sm text-muted-foreground mt-1">{preset.description}</p>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>
      </div>

      {/* Custom Amount */}
      <div className="space-y-4">
        <div className="flex items-center gap-4">
          <div className="flex-1 h-px bg-border" />
          <span className="text-xs font-bold text-muted-foreground tracking-widest uppercase">Or Custom Amount</span>
          <div className="flex-1 h-px bg-border" />
        </div>

        <Card className="border-border/50 shadow-sm">
          <CardContent className="p-6">
            <div className="flex gap-3">
              <div className="flex-1 relative">
                <IndianRupee className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-primary" />
                <Input
                  type="number"
                  placeholder="Enter amount"
                  value={customAmount}
                  onChange={(e) => setCustomAmount(e.target.value)}
                  className="h-14 pl-10 text-lg font-bold bg-secondary/5 border-none rounded-xl"
                  min="1"
                />
              </div>
              <Button
                onClick={handleCustomPay}
                disabled={!razorpayReady}
                className="h-14 px-8 text-base font-bold bg-primary hover:bg-primary/90 rounded-xl shadow-md shadow-primary/20 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <Heart className="mr-2 h-5 w-5" />
                {razorpayReady ? "Pay" : "Loading..."}
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Creators Section */}
      <div className="space-y-4 pt-4">
        <p className="text-xs font-bold text-primary tracking-widest uppercase text-center">The Creators</p>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {developers.map((dev, idx) => (
            <motion.div
              key={dev.name}
              initial={{ opacity: 0, x: idx === 0 ? -20 : 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.3 + idx * 0.15 }}
            >
              <Card className="border-border/50 shadow-sm hover:shadow-md transition-shadow">
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
            </motion.div>
          ))}
        </div>
      </div>

      {/* CTA Banner */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5 }}
      >
        <Card className="overflow-hidden border-none shadow-xl">
          <div className="bg-gradient-to-br from-primary via-orange-500 to-orange-600 p-8 text-center text-white relative overflow-hidden">
            {/* Decorative floating elements */}
            <div className="absolute top-4 left-8 opacity-20">
              <Sparkles className="h-8 w-8" />
            </div>
            <div className="absolute bottom-6 right-10 opacity-20">
              <Coffee className="h-10 w-10" />
            </div>
            <div className="absolute top-12 right-20 opacity-15">
              <Rocket className="h-6 w-6" />
            </div>

            <Heart className="h-10 w-10 mx-auto mb-4 drop-shadow-md" />
            <h3 className="text-2xl font-bold font-lexend mb-2 text-white">Every Contribution Counts</h3>
            <p className="text-white/90 max-w-md mx-auto text-sm leading-relaxed">
              Your support helps us keep the servers running, maintain the app, and build new features. 
              We are students building this for the college community. Thank you! 🙏
            </p>
          </div>
        </Card>
      </motion.div>

      <div className="text-center text-xs text-muted-foreground pb-8">
        <p>Payments are securely processed via Razorpay. No data is stored on our servers.</p>
      </div>
    </div>
    </>
  );
}
