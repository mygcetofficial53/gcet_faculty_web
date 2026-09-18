import { createClient } from "@supabase/supabase-js";

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL;
const supabaseKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY;

// Create client safely so it doesn't crash the entire app if env vars are missing
export const supabase = (supabaseUrl && supabaseKey) 
  ? createClient(supabaseUrl, supabaseKey) 
  : ({
      from: () => ({ upsert: async () => ({ error: new Error("Supabase URL is missing") }) }),
      rpc: async () => ({ error: new Error("Supabase URL is missing") })
    } as any);
