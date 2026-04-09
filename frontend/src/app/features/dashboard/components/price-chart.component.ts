import {
  Component,
  ElementRef,
  inject,
  input,
  OnDestroy,
  OnInit,
  signal,
  ViewChild,
  effect,
} from '@angular/core';
import { FormsModule } from '@angular/forms';
import {
  createChart,
  IChartApi,
  ISeriesApi,
  CandlestickData,
  Time,
} from 'lightweight-charts';
import { ApiService } from '../../../core/api/api.service';
import { WebSocketService } from '../../../core/websocket/websocket.service';

const SYMBOLS = ['AAPL', 'TSLA', 'MSFT', 'GOOGL'];
const INTERVALS = ['1m', '5m', '1h'];

@Component({
  selector: 'tk-price-chart',
  standalone: true,
  imports: [FormsModule],
  template: `
    <div class="card chart-card">
      <div class="card-header">
        <h3>Price Chart</h3>
        <div class="controls">
          <select [(ngModel)]="selectedSymbol" (ngModelChange)="onSymbolChange($event)">
            @for (s of symbols; track s) {
              <option [value]="s">{{ s }}</option>
            }
          </select>
          <select [(ngModel)]="selectedInterval" (ngModelChange)="onIntervalChange($event)">
            @for (i of intervals; track i) {
              <option [value]="i">{{ i }}</option>
            }
          </select>
        </div>
      </div>
      <div #chartContainer class="chart-container"></div>
    </div>
  `,
  styles: [`
    .chart-card { display: flex; flex-direction: column; }
    .card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem; }
    .controls { display: flex; gap: 0.5rem; }
    .chart-container { width: 100%; height: 400px; }
    select { padding: 0.25rem 0.5rem; border-radius: 4px; border: 1px solid #444; background: #1e1e2e; color: #cdd6f4; }
  `],
})
export class PriceChartComponent implements OnInit, OnDestroy {
  @ViewChild('chartContainer', { static: true }) private chartEl!: ElementRef<HTMLDivElement>;

  private readonly api = inject(ApiService);
  private readonly ws = inject(WebSocketService);

  private chart: IChartApi | null = null;
  private series: ISeriesApi<'Candlestick'> | null = null;

  readonly symbols = SYMBOLS;
  readonly intervals = INTERVALS;
  selectedSymbol = 'AAPL';
  selectedInterval = '1m';

  constructor() {
    effect(() => {
      const ticks = this.ws.ticks();
      const tick = ticks.get(this.selectedSymbol);
      if (tick && this.series) {
        // Update last candle with latest price (approximation)
        const now = Math.floor(Date.now() / 1000) as Time;
        this.series.update({
          time: now,
          open: tick.price,
          high: tick.price,
          low: tick.price,
          close: tick.price,
        });
      }
    });
  }

  ngOnInit(): void {
    this.initChart();
    this.loadCandles();
  }

  private initChart(): void {
    this.chart = createChart(this.chartEl.nativeElement, {
      layout: {
        background: { color: '#1e1e2e' },
        textColor: '#cdd6f4',
      },
      grid: {
        vertLines: { color: '#313244' },
        horzLines: { color: '#313244' },
      },
      width: this.chartEl.nativeElement.clientWidth,
      height: 400,
    });

    this.series = this.chart.addCandlestickSeries({
      upColor: '#a6e3a1',
      downColor: '#f38ba8',
      borderVisible: false,
      wickUpColor: '#a6e3a1',
      wickDownColor: '#f38ba8',
    });

    const resizeObserver = new ResizeObserver(() => {
      this.chart?.applyOptions({ width: this.chartEl.nativeElement.clientWidth });
    });
    resizeObserver.observe(this.chartEl.nativeElement);
  }

  private loadCandles(): void {
    this.api.getCandles(this.selectedSymbol, this.selectedInterval).subscribe({
      next: candles => {
        if (this.series) {
          const data: CandlestickData[] = candles.map(c => ({
            time: Math.floor(new Date(c.time).getTime() / 1000) as Time,
            open: c.open,
            high: c.high,
            low: c.low,
            close: c.close,
          }));
          this.series.setData(data);
          this.chart?.timeScale().fitContent();
        }
      },
    });
  }

  onSymbolChange(symbol: string): void {
    this.ws.unsubscribe(`ticks:${this.selectedSymbol}`);
    this.selectedSymbol = symbol;
    this.ws.subscribe(`ticks:${symbol}`);
    this.loadCandles();
  }

  onIntervalChange(_interval: string): void {
    this.loadCandles();
  }

  ngOnDestroy(): void {
    this.chart?.remove();
  }
}
