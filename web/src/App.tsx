import { useEffect, useState, useCallback } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Toaster } from "@/components/ui/sonner";
import { getToken, setToken, getStatus } from "@/lib/api";
import { LoginPage } from "@/components/LoginPage";
import { StatusBar } from "@/components/StatusBar";
import { SubscriptionPanel } from "@/components/SubscriptionPanel";
import { ProxyPanel } from "@/components/ProxyPanel";
import { HealthPanel } from "@/components/HealthPanel";

export default function App() {
  const [authed, setAuthed] = useState(false);
  const [checking, setChecking] = useState(true);

  // 启动时验证已存的 token
  useEffect(() => {
    const token = getToken();
    if (!token) {
      setChecking(false);
      return;
    }
    getStatus()
      .then(() => setAuthed(true))
      .catch(() => setToken(""))
      .finally(() => setChecking(false));
  }, []);

  const handleLogout = useCallback(() => {
    setToken("");
    setAuthed(false);
  }, []);

  if (checking) return null;

  if (!authed) {
    return (
      <>
        <Toaster richColors position="top-right" />
        <LoginPage onLogin={() => setAuthed(true)} />
      </>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <Toaster richColors position="top-right" />

      <header className="border-b">
        <div className="max-w-5xl mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-xl font-bold">Clash 订阅聚合管理</h1>
          <StatusBar onLogout={handleLogout} />
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
