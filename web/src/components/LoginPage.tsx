import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Loader2, Lock } from "lucide-react";
import { setToken, getStatus } from "@/lib/api";

interface Props {
  onLogin: () => void;
}

export function LoginPage({ onLogin }: Props) {
  const [value, setValue] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleLogin = async () => {
    const t = value.trim();
    if (!t) {
      setError("请输入 Token");
      return;
    }
    setLoading(true);
    setError("");

    // 先存 token，再调 API 验证
    setToken(t);
    try {
      await getStatus();
      onLogin();
    } catch {
      setToken("");
      setError("Token 无效或服务不可用");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
            <Lock className="h-6 w-6 text-primary" />
          </div>
          <CardTitle>Clash 订阅聚合管理</CardTitle>
          <p className="text-sm text-muted-foreground mt-1">请输入 API Token 登录</p>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <Label htmlFor="login-token">Token</Label>
              <Input
                id="login-token"
                type="password"
                value={value}
                onChange={(e) => { setValue(e.target.value); setError(""); }}
                placeholder="输入 API Token"
                className="mt-1"
                onKeyDown={(e) => { if (e.key === "Enter") handleLogin(); }}
                autoFocus
              />
              {error && <p className="text-sm text-destructive mt-1">{error}</p>}
            </div>
            <Button className="w-full" onClick={handleLogin} disabled={loading}>
              {loading && <Loader2 className="h-4 w-4 animate-spin mr-2" />}
              登录
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
