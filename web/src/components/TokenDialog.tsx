import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (token: string) => void;
  currentToken: string;
}

export function TokenDialog({ open, onOpenChange, onSave, currentToken }: Props) {
  const [value, setValue] = useState(currentToken);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>设置 API Token</DialogTitle>
        </DialogHeader>
        <div className="py-4">
          <Label htmlFor="token">Token</Label>
          <Input
            id="token"
            type="password"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder="输入 API Token"
            className="mt-2"
            onKeyDown={(e) => {
              if (e.key === "Enter" && value.trim()) onSave(value.trim());
            }}
          />
        </div>
        <DialogFooter>
          <Button onClick={() => value.trim() && onSave(value.trim())}>
            保存
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
