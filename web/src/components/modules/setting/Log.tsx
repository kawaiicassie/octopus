'use client';

import { useEffect, useState, useRef } from 'react';
import { useTranslations } from 'next-intl';
import { ScrollText, Calendar, Trash2, ShieldCheck } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { useSettingList, useSetSetting, SettingKey } from '@/api/endpoints/setting';
import { useClearLogs } from '@/api/endpoints/log';
import { toast } from '@/components/common/Toast';

export function SettingLog() {
    const t = useTranslations('setting');
    const { data: settings } = useSettingList();
    const setSetting = useSetSetting();
    const clearLogs = useClearLogs();

    const [enabled, setEnabled] = useState(true);
    const [saveContent, setSaveContent] = useState(true);
    const [keepPeriod, setKeepPeriod] = useState('7');
    const [isClearing, setIsClearing] = useState(false);

    const initialEnabled = useRef(true);
    const initialSaveContent = useRef(true);
    const initialKeepPeriod = useRef('7');

    useEffect(() => {
        if (settings) {
            const enabledSetting = settings.find(s => s.key === SettingKey.RelayLogKeepEnabled);
            const saveContentSetting = settings.find(s => s.key === SettingKey.RelayLogSaveContent);
            const periodSetting = settings.find(s => s.key === SettingKey.RelayLogKeepPeriod);
            if (enabledSetting) {
                const isEnabled = enabledSetting.value === 'true';
                queueMicrotask(() => setEnabled(isEnabled));
                initialEnabled.current = isEnabled;
            }
            if (saveContentSetting) {
                const isSaveContent = saveContentSetting.value === 'true';
                queueMicrotask(() => setSaveContent(isSaveContent));
                initialSaveContent.current = isSaveContent;
            }
            if (periodSetting) {
                queueMicrotask(() => setKeepPeriod(periodSetting.value));
                initialKeepPeriod.current = periodSetting.value;
            }
        }
    }, [settings]);

    const handleEnabledChange = (checked: boolean) => {
        setEnabled(checked);
        setSetting.mutate(
            { key: SettingKey.RelayLogKeepEnabled, value: checked ? 'true' : 'false' },
            {
                onSuccess: () => {
                    toast.success(t('saved'));
                    initialEnabled.current = checked;
                }
            }
        );
    };

    const handleSaveContentChange = (checked: boolean) => {
        setSaveContent(checked);
        setSetting.mutate(
            { key: SettingKey.RelayLogSaveContent, value: checked ? 'true' : 'false' },
            {
                onSuccess: () => {
                    toast.success(t('saved'));
                    initialSaveContent.current = checked;
                }
            }
        );
    };

    const handleKeepPeriodSave = () => {
        if (keepPeriod === initialKeepPeriod.current) return;

        setSetting.mutate(
            { key: SettingKey.RelayLogKeepPeriod, value: keepPeriod },
            {
                onSuccess: () => {
                    toast.success(t('saved'));
                    initialKeepPeriod.current = keepPeriod;
                }
            }
        );
    };

    const handleClearLogs = () => {
        setIsClearing(true);
        clearLogs.mutate(undefined, {
            onSuccess: () => {
                toast.success(t('log.clearSuccess'));
                setIsClearing(false);
            },
            onError: () => {
                toast.error(t('log.clearFailed'));
                setIsClearing(false);
            }
        });
    };

    return (
        <div className="rounded-3xl border border-border bg-card p-6 custom-shadow space-y-5">
            <h2 className="text-lg font-bold text-card-foreground flex items-center gap-2">
                <ScrollText className="h-5 w-5" />
                {t('log.title')}
            </h2>

            {/* 是否启用历史日志 */}
            <div className="flex items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                    <ScrollText className="h-5 w-5 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('log.enabled.label')}</span>
                </div>
                <Switch
                    checked={enabled}
                    onCheckedChange={handleEnabledChange}
                />
            </div>

            {/* 是否保存请求/响应内容（隐私保护） */}
            <div className="flex items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                    <ShieldCheck className="h-5 w-5 text-muted-foreground" />
                    <div className="flex flex-col">
                        <span className="text-sm font-medium">{t('log.saveContent.label')}</span>
                        <span className="text-xs text-muted-foreground">{t('log.saveContent.hint')}</span>
                    </div>
                </div>
                <Switch
                    checked={saveContent}
                    onCheckedChange={handleSaveContentChange}
                />
            </div>

            {/* 历史日志保存范围 */}
            <div className="flex items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                    <Calendar className="h-5 w-5 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('log.keepPeriod.label')}</span>
                </div>
                <Input
                    type="number"
                    value={keepPeriod}
                    onChange={(e) => setKeepPeriod(e.target.value)}
                    onBlur={handleKeepPeriodSave}
                    placeholder={t('log.keepPeriod.placeholder')}
                    className="w-48 rounded-xl"
                    disabled={!enabled}
                />
            </div>

            {/* 清空历史日志 */}
            <div className="flex items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                    <Trash2 className="h-5 w-5 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('log.clear.label')}</span>
                </div>
                <Button
                    variant="destructive"
                    size="sm"
                    onClick={handleClearLogs}
                    disabled={isClearing}
                    className="rounded-xl"
                >
                    {isClearing ? t('log.clear.clearing') : t('log.clear.button')}
                </Button>
            </div>
        </div>
    );
}

