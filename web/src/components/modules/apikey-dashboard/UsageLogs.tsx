'use client';

import { useMemo, useCallback, useRef, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import { useAPIKeyLogs, SimpleLog } from '@/api/endpoints/apikey';
import { ScrollText, ChevronDown, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import dayjs from 'dayjs';

export function UsageLogs() {
    const t = useTranslations('apiKeyDashboard');
    const { data, hasNextPage, isFetchingNextPage, fetchNextPage, isLoading } = useAPIKeyLogs({ pageSize: 10 });
    const loadMoreRef = useRef<HTMLDivElement>(null);

    // Flatten pages into a single array
    const logs = useMemo(() => {
        const pages = data?.pages ?? [];
        const seen = new Set<number>();
        const merged: SimpleLog[] = [];

        for (const page of pages) {
            for (const log of page) {
                if (seen.has(log.id)) continue;
                seen.add(log.id);
                merged.push(log);
            }
        }

        merged.sort((a, b) => b.time - a.time);
        return merged;
    }, [data]);

    const loadMore = useCallback(() => {
        if (hasNextPage && !isFetchingNextPage) {
            void fetchNextPage();
        }
    }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

    // Intersection observer for infinite scroll
    useEffect(() => {
        const observer = new IntersectionObserver(
            (entries) => {
                if (entries[0].isIntersecting) {
                    loadMore();
                }
            },
            { threshold: 0.1 }
        );

        if (loadMoreRef.current) {
            observer.observe(loadMoreRef.current);
        }

        return () => observer.disconnect();
    }, [loadMore]);

    const formatCost = (cost: number) => {
        if (cost >= 1) return `$${cost.toFixed(4)}`;
        if (cost >= 0.001) return `$${cost.toFixed(6)}`;
        return `$${cost.toFixed(8)}`;
    };

    const formatTokens = (tokens: number) => {
        if (tokens >= 1000000) return `${(tokens / 1000000).toFixed(1)}M`;
        if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}K`;
        return tokens.toString();
    };

    return (
        <div className="custom-shadow rounded-2xl border bg-card p-6">
            <div className="flex items-center gap-2 mb-4">
                <ScrollText className="w-5 h-5 text-chart-5" />
                <span className="font-semibold">{t('usageLogs.title')}</span>
            </div>

            {isLoading ? (
                <div className="flex items-center justify-center py-12">
                    <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
                </div>
            ) : logs.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                    {t('usageLogs.noLogs')}
                </div>
            ) : (
                <>
                    {/* Table */}
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead>
                                <tr className="border-b border-border/50">
                                    <th className="text-left py-3 px-2 text-muted-foreground font-medium">{t('usageLogs.date')}</th>
                                    <th className="text-left py-3 px-2 text-muted-foreground font-medium">{t('usageLogs.model')}</th>
                                    <th className="text-right py-3 px-2 text-muted-foreground font-medium whitespace-nowrap">{t('usageLogs.inputTokens')}</th>
                                    <th className="text-right py-3 px-2 text-muted-foreground font-medium whitespace-nowrap">{t('usageLogs.outputTokens')}</th>
                                    <th className="text-right py-3 px-2 text-muted-foreground font-medium whitespace-nowrap">{t('usageLogs.totalTokens')}</th>
                                    <th className="text-right py-3 px-2 text-muted-foreground font-medium">{t('usageLogs.cost')}</th>
                                </tr>
                            </thead>
                            <tbody>
                                {logs.map((log) => (
                                    <tr key={log.id} className="border-b border-border/30 hover:bg-muted/30 transition-colors">
                                        <td className="py-3 px-2 whitespace-nowrap text-muted-foreground">
                                            {dayjs.unix(log.time).format('MM-DD HH:mm')}
                                        </td>
                                        <td className="py-3 px-2">
                                            <span className="inline-flex items-center px-2 py-0.5 rounded-md bg-secondary text-secondary-foreground text-xs font-medium truncate max-w-[200px]">
                                                {log.model_name}
                                            </span>
                                        </td>
                                        <td className="py-3 px-2 text-right tabular-nums">
                                            {formatTokens(log.input_tokens)}
                                        </td>
                                        <td className="py-3 px-2 text-right tabular-nums">
                                            {formatTokens(log.output_tokens)}
                                        </td>
                                        <td className="py-3 px-2 text-right tabular-nums font-medium">
                                            {formatTokens(log.total_tokens)}
                                        </td>
                                        <td className="py-3 px-2 text-right tabular-nums text-chart-1 font-medium">
                                            {formatCost(log.cost)}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>

                    {/* Load more trigger */}
                    <div ref={loadMoreRef} className="pt-4 flex justify-center">
                        {isFetchingNextPage ? (
                            <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
                        ) : hasNextPage ? (
                            <Button
                                variant="ghost"
                                size="sm"
                                onClick={loadMore}
                                className="text-muted-foreground hover:text-foreground"
                            >
                                <ChevronDown className="w-4 h-4 mr-1" />
                                {t('usageLogs.loadMore')}
                            </Button>
                        ) : logs.length > 0 ? (
                            <span className="text-xs text-muted-foreground">{t('usageLogs.noMore')}</span>
                        ) : null}
                    </div>
                </>
            )}
        </div>
    );
}
