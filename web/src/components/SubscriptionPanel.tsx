import { useState, useEffect, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Plus, Trash2, RefreshCw, Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  listSubscriptions,
  addSubscription,
  deleteSubscription,
  refreshAllSubscriptions,
  refreshOneSubscription,
  type SubInfo,
} from "@/lib/api";

export function SubscriptionPanel() {
  const [subs, setSubs] = useState<SubInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshingId, setRefreshingId] = useState<string | null>(null);
  const [refreshingAll, setRefreshingAll] = useState(false);
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [adding, setAdding] = useState(false);

  const load = useCallback(() => {
    setLoading(true);
    listSubscriptions()
      .then((r) => setSubs(r.subscriptions || []))
      .catch((e) => toast.error(e.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleAdd = async () => {
    if (!name.trim() || !url.trim()) {
      toast.error("名称和 URL 不能为空");
      return;
    }
    setAdding(true);
    try {
      const r = await addSubscription(name.trim(), url.trim());
      toast.success(`${r.message}，${r.proxy_count} 个节点`);
      setName("");
      setUrl("");
      load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "添加失败");
    } finally {
      setAdding(false);
    }
  };

  const handleDelete = async (id: string, subName: string) => {
    if (!confirm(`确定删除订阅「${subName}」？`)) return;
    try {
      await deleteSubscription(id);
      toast.success("已删除");
      load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "删除失败");
    }
  };

  const handleRefreshAll = async () => {
    setRefreshingAll(true);
    try {
      const r = await refreshAllSubscriptions();
      toast.success(`${r.message}，共 ${r.proxy_count} 个节点`);
      load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "刷新失败");
    } finally {
      setRefreshingAll(false);
    }
  };

  const handleRefreshOne = async (id: string) => {
    setRefreshingId(id);
    try {
      const r = await refreshOneSubscription(id);
      toast.success(`${r.message}，${r.proxy_count} 个节点`);
      load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "刷新失败");
    } finally {
      setRefreshingId(null);
    }
  };

  const totalProxies = subs.reduce((s, sub) => s + sub.proxy_count, 0);

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">添加订阅</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-[1fr_2fr_auto] gap-3 items-end">
            <div>
              <Label htmlFor="sub-name">名称</Label>
              <Input
                id="sub-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="机场名称"
                className="mt-1"
              />
            </div>
            <div>
              <Label htmlFor="sub-url">订阅 URL</Label>
              <Input
                id="sub-url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://..."
                className="mt-1"
              />
            </div>
            <Button onClick={handleAdd} disabled={adding}>
              {adding ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
              <span className="ml-1">添加</span>
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-base">
              订阅列表
              <Badge variant="secondary" className="ml-2">{subs.length} 个订阅</Badge>
              <Badge variant="outline" className="ml-2">{totalProxies} 个节点</Badge>
            </CardTitle>
          </div>
          <Button variant="outline" size="sm" onClick={handleRefreshAll} disabled={refreshingAll}>
            {refreshingAll ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
            <span className="ml-1">全部刷新</span>
          </Button>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : subs.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">暂无订阅，请先添加</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>名称</TableHead>
                  <TableHead>节点数</TableHead>
                  <TableHead>更新时间</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {subs.map((sub) => (
                  <TableRow key={sub.id}>
                    <TableCell className="font-medium">{sub.name}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{sub.proxy_count}</Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {sub.updated_at}
                    </TableCell>
                    <TableCell className="text-right space-x-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleRefreshOne(sub.id)}
                        disabled={refreshingId === sub.id}
                        aria-label="刷新"
                      >
                        {refreshingId === sub.id ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <RefreshCw className="h-4 w-4" />
                        )}
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(sub.id, sub.name)}
                        aria-label="删除"
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
