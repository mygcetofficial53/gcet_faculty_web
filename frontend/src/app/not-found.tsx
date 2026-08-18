import Link from "next/link";
import { Button } from "@/components/ui/button";
import { GraduationCap } from "lucide-react";

export default function NotFound() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-background p-4 text-center">
      <div className="bg-primary/10 p-6 rounded-full mb-6">
        <GraduationCap className="h-20 w-20 text-primary" />
      </div>
      <h1 className="text-6xl font-bold font-lora text-secondary mb-4">404</h1>
      <h2 className="text-2xl font-semibold text-secondary mb-2">Page Not Found</h2>
      <p className="text-muted-foreground max-w-md mb-8">
        The page you are looking for doesn't exist or has been moved.
      </p>
      <Link href="/">
        <Button className="h-12 px-8 text-base shadow-lg shadow-primary/25">
          Return to Dashboard
        </Button>
      </Link>
    </div>
  );
}
