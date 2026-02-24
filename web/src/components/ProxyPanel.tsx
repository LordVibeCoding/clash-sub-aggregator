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
import { Zap, Check, Loader2, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import {
  listProxies,
  switchProxy,
  testDelay,
  type ProxyGroup,
} from "@/lib/api";

export function ProxyPanel() {
  const [group, setGroup] = useState<ProxyGroup | null>(null);
  const [delays, setDelays] = useState<Record<string, number | "timeout" | "testing">>(
    {}
  );
  const [switching, setSwitching] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const load = useCallback(() => {
    setLoading(true);
    listProxies()
      .then((r) => {
        const proxies = r.proxies;
        const proxyGroup = proxies["PROXY"] || proxies["proxy"];
        if (proxyGroup) {
          setGroup(proxyGroup);
        } else {
          const groups = Object.values(proxies).filter(
            (g) => g.type === "Selector" || g.type === "URLTest"
          );
          if (groups.length > 0) setGroup(groups[0]);
        }
      })
      .catch((e) => toast.error(e.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleSwitch = async (name: string) => {
    if (!group) return;
    setSwitching(name);
    try {
      const r = await switchProxy(group.name, name);
      toast.success(r.message);
      load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "切换失败");
    } finally {
      setSwitching(null);
    }
  };

  const handleTestDelay = async (name: string) => {
    setDelays((d) => ({ ...d, [name]: "testing" }));
    try {
      const r = await testDelay(name);
      if (r.delay !== undefined) {
        setDelays((d) => ({ ...d, [name]: r.delay as number }));
      } else {
        setDelays((d) => ({ ...d, [name]: "timeout" }));
      }
    } catch {
      setDelays((d) => ({ ...d, [name]: "timeout" }));
    }
  };

  const handleTestAll = async () => {
    if (!group) return;
    for (const name of group.all) {
      handleTestDelay(name);
      await new Promise((r) => setTimeout(r, 200));
    }
  };

  const delayBadge = (name: string) => {
    const d = delays[name];
    if (d === undefined) return null;
    if (d === "testing")
      return <Loader2 className="h-3 w-3 animate-spin text-muted-foreground" />;
    if (d === "timeout")
      return <Badge variant="destructive" className="text-xs">超时</Badge>;
    const ms = d as number;
    const variant = ms < 200 ? "default" : ms < 500 ? "secondary" : "destructive";
    return <Badge variant={variant} className="text-xs">{ms}ms</Badge>;
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle className="text-base">
            代理节点
            {group && (
              <>
                <Badge variant="secondary" className="ml-2">
                  {group.all.length} 个
                </Badge>
                <Badge variant="outline" className="ml-2">
                  当前: {group.now}
                </Badge>
              </>
            )}
          </CardTitle>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={handleTestAll} disabled={!group}>
            <Zap className="h-4 w-4 mr-1" />
            全部测速
          </Button>
          <Button variant="outline" size="sm" onClick={load}>
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : !group ? (
          <p className="text-center text-muted-foreground py-8">
            暂无节点数据
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>节点名称</TableHead>
                <TableHead>延迟</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {group.all.map((name) => (
                <TableRow
                  key={name}
                  className={name === group.now ? "bg-accent/50" : ""}
                >
                  <TableCell className="font-medium">
                    <div className="flex items-center gap-2">
                      {name === group.now && (
                        <Check className="h-4 w-4 text-green-500" />
                      )}
                      {name}
                    </div>
                  </TableCell>
                  <TableCell>{delayBadge(name)}</TableCell>
                  <TableCell className="text-right space-x-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleTestDelay(name)}
                      disabled={delays[name] === "testing"}
                      aria-label="测速"
                    >
                      <Zap className="h-4 w-4" />
                    </Button>
                    {name !== group.now && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleSwitch(name)}
                        disabled={switching === name}
                      >
                        {switching === name ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          "切换"
                        )}
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
