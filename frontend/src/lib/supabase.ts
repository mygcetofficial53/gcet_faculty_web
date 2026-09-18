import { createClient } from "@supabase/supabase-js";

const supabaseUrl = (process.env.NEXT_PUBLIC_SUPABASE_URL || "").trim();
const supabaseKey = (process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY || "").trim();

const isConfigured = supabaseUrl.length > 10 && supabaseKey.length > 10;

// Create client safely so it doesn't crash the entire app if env vars are missing
export const supabase = isConfigured
  ? createClient(supabaseUrl, supabaseKey) 
  : ({
      from: () => ({ upsert: async () => ({ error: new Error("Supabase URL is missing or invalid") }) }),
      rpc: async () => ({ error: new Error("Supabase URL is missing or invalid") })
    } as any);
