import { useState, useEffect } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Settings } from "lucide-react";
import { getStatus, type ServiceStatus } from "@/lib/api";

interface Props {
  onTokenClick: () => void;
}

export function StatusBar({ onTokenClick }: Props) {
  const [status, setStatus] = useState<ServiceStatus | null>(null);

  useEffect(() => {
    const load = () => {
      getStatus().then(setStatus).catch(() => setStatus(null));
    };
    load();
    const id = setInterval(load, 10000);
    return () => clearInterval(id);
  }, []);

  return (
    <div className="flex items-center gap-3">
      {status ? (
        <Badge variant={status.mihomo_running ? "default" : "destructive"}>
          {status.mihomo_running ? "mihomo 运行中" : "mihomo 已停止"}
        </Badge>
      ) : (
        <Badge variant="outline">连接中...</Badge>
      )}
      <Button variant="ghost" size="icon" onClick={onTokenClick} aria-label="设置 Token">
        <Settings className="h-4 w-4" />
      </Button>
    </div>
  );
}
