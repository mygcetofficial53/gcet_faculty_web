"use client";

import { useEffect, useState, useCallback } from "react";
import { api } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import {
  Shield,
  Activity,
  Wifi,
  WifiOff,
  Zap,
  Timer,
  RefreshCw,
  Server,
  TrendingUp,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Loader2,
  Globe2,
  Search,
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

interface ProxyInfo {
  url: string;
  successes: number;
  failures: number;
  avg_ms: number;
  success_rate: number;
  status: "healthy" | "degraded" | "dead";
}

interface PoolStats {
  total_proxies: number;
  healthy_count: number;
  degraded_count: number;
  dead_count: number;
  avg_latency_ms: number;
  top_proxies: ProxyInfo[];
}

interface DiscoveryStats {
  is_running: boolean;
  last_discovery_time: string;
  last_discovery_count: number;
  total_discovered: number;
  total_healthy: number;
}

interface ProxyStatusData {
  pool: PoolStats | null;
  discovery: DiscoveryStats;
}

function maskIP(proxyURL: string): string {
  try {
    const match = proxyURL.match(/\/\/(\d+\.\d+\.\d+\.)\d+/);
    if (match) {
      return proxyURL.replace(match[1], match[1].split(".").slice(0, 2).join(".") + ".***.");
    }
    return proxyURL;
  } catch {
    return proxyURL;
  }
}

function getStatusColor(status: string) {
  switch (status) {
    case "healthy":
      return "bg-emerald-500/10 text-emerald-600 border-emerald-500/20";
    case "degraded":
      return "bg-amber-500/10 text-amber-600 border-amber-500/20";
    case "dead":
      return "bg-red-500/10 text-red-600 border-red-500/20";
    default:
      return "bg-muted text-muted-foreground";
  }
}

function getStatusIcon(status: string) {
  switch (status) {
    case "healthy":
      return <CheckCircle2 className="h-3.5 w-3.5" />;
    case "degraded":
      return <AlertTriangle className="h-3.5 w-3.5" />;
    case "dead":
      return <XCircle className="h-3.5 w-3.5" />;
    default:
      return null;
  }
}

function formatTimeAgo(dateStr: string): string {
  if (!dateStr || dateStr === "0001-01-01T00:00:00Z") return "Never";
  const date = new Date(dateStr);
  const now = new Date();
  const diff = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

export default function ProxyStatusPage() {
  const [data, setData] = useState<ProxyStatusData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);

  const fetchData = useCallback(async (isManual = false) => {
    if (isManual) setRefreshing(true);
    try {
      const res = await api.get("/proxy/status");
      if (res.data.success) {
        setData(res.data.data as ProxyStatusData);
        setError(null);
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to fetch proxy status");
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    if (!autoRefresh) return;
    const interval = setInterval(() => fetchData(), 30000);
    return () => clearInterval(interval);
  }, [autoRefresh, fetchData]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-[60vh]">
        <div className="text-center space-y-4">
          <Loader2 className="h-10 w-10 animate-spin text-primary mx-auto" />
          <p className="text-muted-foreground">Loading proxy status...</p>
        </div>
      </div>
    );
  }

  const pool = data?.pool;
  const discovery = data?.discovery;

  const totalProxies = pool?.total_proxies ?? 0;
  const healthyCount = pool?.healthy_count ?? 0;
  const degradedCount = pool?.degraded_count ?? 0;
  const deadCount = pool?.dead_count ?? 0;
  const avgLatency = pool?.avg_latency_ms ?? 0;
  const healthPercentage = totalProxies > 0 ? Math.round((healthyCount / totalProxies) * 100) : 0;

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ type: "spring", duration: 0.6 }}
            className="w-14 h-14 bg-gradient-to-br from-primary to-orange-600 rounded-2xl flex items-center justify-center shadow-lg shadow-primary/30"
          >
            <Shield className="h-7 w-7 text-white" />
          </motion.div>
          <div>
            <h2 className="text-2xl font-bold font-lora text-secondary tracking-tight">
              Indian Proxy Tunnel
            </h2>
            <p className="text-sm text-muted-foreground">
              Auto-discovering 🇮🇳 Indian proxies to bypass GMS geo-blocking
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            className={autoRefresh ? "border-primary/30 text-primary" : ""}
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            {autoRefresh ? (
              <Activity className="h-4 w-4 mr-1.5 animate-pulse" />
            ) : (
              <WifiOff className="h-4 w-4 mr-1.5" />
            )}
            {autoRefresh ? "Live" : "Paused"}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => fetchData(true)}
            disabled={refreshing}
          >
            <RefreshCw className={`h-4 w-4 mr-1.5 ${refreshing ? "animate-spin" : ""}`} />
            Refresh
          </Button>
        </div>
      </div>

      {error && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="p-4 rounded-xl bg-destructive/10 text-destructive text-sm flex items-center gap-2"
        >
          <AlertTriangle className="h-4 w-4" />
          {error}
        </motion.div>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {[
          {
            title: "Total Proxies",
            value: totalProxies,
            icon: Server,
            color: "text-primary",
            bg: "bg-primary/10",
            subtitle: `${discovery?.total_discovered ?? 0} discovered`,
          },
          {
            title: "Healthy",
            value: healthyCount,
            icon: CheckCircle2,
            color: "text-emerald-600",
            bg: "bg-emerald-500/10",
            subtitle: `${healthPercentage}% success rate`,
          },
          {
            title: "Degraded",
            value: degradedCount,
            icon: AlertTriangle,
            color: "text-amber-600",
            bg: "bg-amber-500/10",
            subtitle: `${deadCount} dead`,
          },
          {
            title: "Avg Latency",
            value: `${avgLatency}ms`,
            icon: Zap,
            color: avgLatency < 2000 ? "text-emerald-600" : avgLatency < 5000 ? "text-amber-600" : "text-red-600",
            bg: avgLatency < 2000 ? "bg-emerald-500/10" : avgLatency < 5000 ? "bg-amber-500/10" : "bg-red-500/10",
            subtitle: "via proxy",
          },
        ].map((stat, idx) => (
          <motion.div
            key={stat.title}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: idx * 0.1 }}
          >
            <Card className="border-border/50 shadow-sm hover:shadow-md transition-shadow">
              <CardContent className="p-5">
                <div className="flex items-center justify-between mb-3">
                  <div className={`h-10 w-10 ${stat.bg} rounded-xl flex items-center justify-center`}>
                    <stat.icon className={`h-5 w-5 ${stat.color}`} />
                  </div>
                  <TrendingUp className="h-4 w-4 text-muted-foreground/30" />
                </div>
                <p className="text-2xl font-black text-secondary">{stat.value}</p>
                <p className="text-xs text-muted-foreground font-medium mt-1">{stat.title}</p>
                <p className="text-[10px] text-muted-foreground/60 mt-0.5">{stat.subtitle}</p>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>

      {/* Discovery Engine Status */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
      >
        <Card className="overflow-hidden border-none shadow-xl">
          <div className="bg-gradient-to-br from-primary via-orange-500 to-orange-600 p-6 text-white relative overflow-hidden">
            {/* Floating decorations */}
            <div className="absolute top-3 right-8 opacity-20">
              <Globe2 className="h-12 w-12" />
            </div>
            <div className="absolute bottom-2 left-10 opacity-10">
              <Search className="h-16 w-16" />
            </div>

            <div className="relative z-10">
              <div className="flex items-center gap-3 mb-4">
                <div className="h-10 w-10 bg-white/20 backdrop-blur-sm rounded-xl flex items-center justify-center">
                  {discovery?.is_running ? (
                    <Loader2 className="h-5 w-5 text-white animate-spin" />
                  ) : (
                    <Wifi className="h-5 w-5 text-white" />
                  )}
                </div>
                <div>
                  <h3 className="font-bold text-lg">Auto-Discovery Engine</h3>
                  <p className="text-white/70 text-xs">
                    Scrapes ProxyScrape, GeoNode & FreeProxyList for 🇮🇳 Indian proxies
                  </p>
                </div>
                <Badge className={`ml-auto ${discovery?.is_running ? "bg-white/20 text-white border-white/30" : "bg-white/10 text-white/80 border-white/20"} backdrop-blur-sm`}>
                  {discovery?.is_running ? "🔍 Scanning..." : "✅ Idle"}
                </Badge>
              </div>

              <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mt-4">
                <div className="bg-white/10 backdrop-blur-sm rounded-xl p-3">
                  <p className="text-white/60 text-[10px] font-medium uppercase tracking-wider">Last Scan</p>
                  <p className="text-white font-bold text-sm mt-1">
                    {formatTimeAgo(discovery?.last_discovery_time ?? "")}
                  </p>
                </div>
                <div className="bg-white/10 backdrop-blur-sm rounded-xl p-3">
                  <p className="text-white/60 text-[10px] font-medium uppercase tracking-wider">Found</p>
                  <p className="text-white font-bold text-sm mt-1">
                    {discovery?.last_discovery_count ?? 0} proxies
                  </p>
                </div>
                <div className="bg-white/10 backdrop-blur-sm rounded-xl p-3">
                  <p className="text-white/60 text-[10px] font-medium uppercase tracking-wider">GMS Validated</p>
                  <p className="text-white font-bold text-sm mt-1">
                    {discovery?.total_healthy ?? 0} healthy
                  </p>
                </div>
                <div className="bg-white/10 backdrop-blur-sm rounded-xl p-3">
                  <p className="text-white/60 text-[10px] font-medium uppercase tracking-wider">Next Scan</p>
                  <p className="text-white font-bold text-sm mt-1">
                    <Timer className="h-3.5 w-3.5 inline mr-1" />
                    ~15 min cycle
                  </p>
                </div>
              </div>
            </div>
          </div>
        </Card>
      </motion.div>

      {/* Pool Health Bar */}
      {totalProxies > 0 && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
        >
          <Card className="border-border/50 shadow-sm">
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                Pool Health Distribution
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex rounded-full overflow-hidden h-4 bg-muted/30">
                {healthyCount > 0 && (
                  <motion.div
                    initial={{ width: 0 }}
                    animate={{ width: `${(healthyCount / totalProxies) * 100}%` }}
                    transition={{ duration: 0.8, delay: 0.6 }}
                    className="bg-emerald-500 h-full"
                    title={`${healthyCount} healthy`}
                  />
                )}
                {degradedCount > 0 && (
                  <motion.div
                    initial={{ width: 0 }}
                    animate={{ width: `${(degradedCount / totalProxies) * 100}%` }}
                    transition={{ duration: 0.8, delay: 0.8 }}
                    className="bg-amber-500 h-full"
                    title={`${degradedCount} degraded`}
                  />
                )}
                {deadCount > 0 && (
                  <motion.div
                    initial={{ width: 0 }}
                    animate={{ width: `${(deadCount / totalProxies) * 100}%` }}
                    transition={{ duration: 0.8, delay: 1.0 }}
                    className="bg-red-500 h-full"
                    title={`${deadCount} dead`}
                  />
                )}
              </div>
              <div className="flex items-center gap-6 mt-3 text-xs text-muted-foreground">
                <span className="flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-emerald-500" />
                  Healthy ({healthyCount})
                </span>
                <span className="flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-amber-500" />
                  Degraded ({degradedCount})
                </span>
                <span className="flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-red-500" />
                  Dead ({deadCount})
                </span>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      )}

      {/* Top Proxies Table */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6 }}
      >
        <Card className="border-border/50 shadow-sm">
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                Active Proxy Pool
              </CardTitle>
              <Badge variant="secondary" className="text-xs">
                Top {pool?.top_proxies?.length ?? 0}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            {pool?.top_proxies && pool.top_proxies.length > 0 ? (
              <div className="rounded-xl border border-border/50 overflow-hidden">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-muted/30 hover:bg-muted/30">
                      <TableHead className="font-semibold text-xs">Proxy</TableHead>
                      <TableHead className="font-semibold text-xs text-center">Status</TableHead>
                      <TableHead className="font-semibold text-xs text-center">Wins</TableHead>
                      <TableHead className="font-semibold text-xs text-center">Fails</TableHead>
                      <TableHead className="font-semibold text-xs text-center">Rate</TableHead>
                      <TableHead className="font-semibold text-xs text-center">Latency</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    <AnimatePresence>
                      {pool.top_proxies.map((proxy, idx) => (
                        <motion.tr
                          key={proxy.url}
                          initial={{ opacity: 0, x: -20 }}
                          animate={{ opacity: 1, x: 0 }}
                          transition={{ delay: 0.7 + idx * 0.05 }}
                          className="border-b border-border/30 hover:bg-muted/20 transition-colors"
                        >
                          <TableCell className="font-mono text-xs text-muted-foreground">
                            {maskIP(proxy.url)}
                          </TableCell>
                          <TableCell className="text-center">
                            <Badge
                              variant="outline"
                              className={`text-[10px] px-2 py-0.5 gap-1 ${getStatusColor(proxy.status)}`}
                            >
                              {getStatusIcon(proxy.status)}
                              {proxy.status}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-center text-xs font-semibold text-emerald-600">
                            {proxy.successes}
                          </TableCell>
                          <TableCell className="text-center text-xs font-semibold text-red-500">
                            {proxy.failures}
                          </TableCell>
                          <TableCell className="text-center">
                            <span className={`text-xs font-bold ${
                              proxy.success_rate >= 0.7
                                ? "text-emerald-600"
                                : proxy.success_rate >= 0.3
                                  ? "text-amber-600"
                                  : "text-red-600"
                            }`}>
                              {(proxy.success_rate * 100).toFixed(0)}%
                            </span>
                          </TableCell>
                          <TableCell className="text-center text-xs text-muted-foreground">
                            {proxy.avg_ms > 0 ? `${proxy.avg_ms}ms` : "—"}
                          </TableCell>
                        </motion.tr>
                      ))}
                    </AnimatePresence>
                  </TableBody>
                </Table>
              </div>
            ) : (
              <div className="text-center py-12 text-muted-foreground">
                <Server className="h-12 w-12 mx-auto mb-4 opacity-30" />
                <p className="font-medium">No proxies in pool yet</p>
                <p className="text-sm mt-1">The auto-discovery engine will populate this shortly...</p>
              </div>
            )}
          </CardContent>
        </Card>
      </motion.div>

      {/* Info Banner */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.8 }}
      >
        <Card className="border-border/50 bg-muted/30">
          <CardContent className="p-5">
            <div className="flex items-start gap-3">
              <div className="h-8 w-8 bg-primary/10 rounded-lg flex items-center justify-center shrink-0 mt-0.5">
                <Globe2 className="h-4 w-4 text-primary" />
              </div>
              <div className="text-sm text-muted-foreground leading-relaxed">
                <p className="font-semibold text-secondary mb-1">How it works</p>
                <p>
                  The GCET GMS Portal blocks non-Indian IPs. This engine automatically discovers free Indian
                  HTTP/SOCKS5 proxies from multiple public sources, validates each one against the GMS login page,
                  and feeds healthy proxies into the racing pool. When your app makes a request to GMS,
                  5 proxies race simultaneously — the fastest response wins.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
