import { useState, useEffect, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ShieldCheck, Loader2, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { getHealth, triggerHealthCheck, type HealthStatus } from "@/lib/api";

export function HealthPanel() {
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [triggering, setTriggering] = useState(false);

  const load = useCallback(() => {
    setLoading(true);
    getHealth()
      .then(setHealth)
      .catch((e) => toast.error(e.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => { load(); }, [load]);

  // 检查中时自动轮询
  useEffect(() => {
    if (!health?.checking) return;
    const id = setInterval(load, 3000);
    return () => clearInterval(id);
  }, [health?.checking, load]);

  const handleTrigger = async () => {
    setTriggering(true);
    try {
      const r = await triggerHealthCheck();
      toast.success(r.message);
      // 延迟一下再刷新状态
      setTimeout(load, 1000);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "触发失败");
    } finally {
      setTriggering(false);
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base flex items-center gap-2">
            <ShieldCheck className="h-5 w-5" />
            健康检查
            {health?.checking && (
              <Badge variant="secondary" className="ml-2">
                <Loader2 className="h-3 w-3 animate-spin mr-1" />
                检查中...
              </Badge>
            )}
          </CardTitle>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleTrigger}
              disabled={triggering || health?.checking}
            >
              {triggering ? (
                <Loader2 className="h-4 w-4 animate-spin mr-1" />
              ) : (
                <ShieldCheck className="h-4 w-4 mr-1" />
              )}
              触发检查
            </Button>
            <Button variant="outline" size="sm" onClick={load}>
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {loading && !health ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : !health ? (
            <p className="text-center text-muted-foreground py-8">无法获取健康状态</p>
          ) : (
            <div className="space-y-4">
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
                <div className="rounded-lg border p-3">
                  <p className="text-sm text-muted-foreground">黑名单节点</p>
                  <p className="text-2xl font-bold">{health.blacklist_count}</p>
                </div>
                <div className="rounded-lg border p-3">
                  <p className="text-sm text-muted-foreground">上次检查</p>
                  <p className="text-sm font-medium mt-1">
                    {health.last_check_at || "从未检查"}
                  </p>
                </div>
                {health.last_check_cost && (
                  <div className="rounded-lg border p-3">
                    <p className="text-sm text-muted-foreground">检查耗时</p>
                    <p className="text-sm font-medium mt-1">{health.last_check_cost}</p>
                  </div>
                )}
              </div>

              {health.blacklist.length > 0 && (
                <div>
                  <h3 className="text-sm font-medium mb-2">黑名单</h3>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>节点名称</TableHead>
                        <TableHead>加入时间</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {health.blacklist.map((item) => (
                        <TableRow key={item.name}>
                          <TableCell>
                            <Badge variant="destructive" className="font-normal">
                              {item.name}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-muted-foreground text-sm">
                            {item.since}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
