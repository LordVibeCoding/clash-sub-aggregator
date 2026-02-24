import { useEffect, useCallback, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Toaster } from "@/components/ui/sonner";
import { toast } from "sonner";
import { getToken, setToken } from "@/lib/api";
import { TokenDialog } from "@/components/TokenDialog";
import { StatusBar } from "@/components/StatusBar";
import { SubscriptionPanel } from "@/components/SubscriptionPanel";
import { ProxyPanel } from "@/components/ProxyPanel";
import { HealthPanel } from "@/components/HealthPanel";

export default function App() {
  const [token, setTokenState] = useState(getToken());
  const [showTokenDialog, setShowTokenDialog] = useState(!token);

  const handleSetToken = useCallback((t: string) => {
    setToken(t);
    setTokenState(t);
    setShowTokenDialog(false);
    toast.success("Token 已保存");
  }, []);

  useEffect(() => {
    if (!token) setShowTokenDialog(true);
  }, [token]);

  return (
    <div className="min-h-screen bg-background">
      <Toaster richColors position="top-right" />
      <TokenDialog
        open={showTokenDialog}
        onOpenChange={setShowTokenDialog}
        onSave={handleSetToken}
        currentToken={token}
      />

      <header className="border-b">
        <div className="max-w-5xl mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-xl font-bold">Clash 订阅聚合管理</h1>
          <StatusBar onTokenClick={() => setShowTokenDialog(true)} />
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 py-6">
        <Tabs defaultValue="subscriptions">
          <TabsList className="mb-4">
            <TabsTrigger value="subscriptions">订阅管理</TabsTrigger>
            <TabsTrigger value="proxies">节点列表</TabsTrigger>
            <TabsTrigger value="health">健康检查</TabsTrigger>
          </TabsList>

          <TabsContent value="subscriptions">
            <SubscriptionPanel />
          </TabsContent>
          <TabsContent value="proxies">
            <ProxyPanel />
          </TabsContent>
          <TabsContent value="health">
            <HealthPanel />
          </TabsContent>
        </Tabs>
      </main>
    </div>
  );
}
