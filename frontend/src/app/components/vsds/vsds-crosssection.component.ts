import {
  Component, OnInit, OnDestroy,
  ViewChild, ElementRef, ChangeDetectorRef,
} from '@angular/core';
import { FormsModule }      from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { AuthService }      from '../../auth/auth.service';
import {
  BundlesService, Bundle,
}                           from '../../services/bundles.service';
import {
  VsdsService,
  SectorVoxel,
}                           from '../../services/vsds.service';
import {
  galaxyMapUrl, galaxyMapViewBox,
}                           from '../../utils/galaxy-map';
import { fetchBundle }      from '../../utils/bundle-loader';
import {
  colorScale, VIRIDIS, rhoLogT, rhoToHex,
}                           from '../../utils/color-scale';
import * as THREE           from 'three';
import { OrbitControls }
  from 'three/examples/jsm/controls/OrbitControls.js';
import { SelectModule }         from 'primeng/select';
import { SelectButtonModule }   from 'primeng/selectbutton';
import { InputNumberModule }    from 'primeng/inputnumber';
import { ButtonModule }         from 'primeng/button';
import { MessageModule }        from 'primeng/message';

// ── Types ──────────────────────────────────────────────────────────────────

interface BundleSurveyPoint {
  sysname:     string;
  zsample:     number;
  x:  number; y: number; z: number;
  gc_x: number; gc_y: number; gc_z: number;
  syscount:    number;
  maxdistance: number;
  rho:         number;
}

type SelectionState = 'idle' | 'a_set' | 'ready';

interface GalacticPoint { gc_x: number; gc_z: number; }

interface RhoScale {
  dataMin:  number;
  dataMax:  number;
  scaleMin: number;
  scaleMax: number;
}

interface CoverageCell {
  gc_x:  number;
  gc_z:  number;
  count: number;
}

interface BinCell {
  xBin:    number;
  along:   number;
  zsample: number;
  avgRho:  number;
}

interface BinResult {
  cells:         BinCell[];
  abLength:      number;
  unitAB:        { x: number; z: number };
  yValues:       number[];
  xBinMin:       number;
  xBinMax:       number;
  filteredCount: number;
}

interface SelectOption { label: string; value: string; }

// ── Component ──────────────────────────────────────────────────────────────

@Component({
  selector:    'app-vsds-crosssection',
  standalone:  true,
  imports: [
    FormsModule,
    SelectModule,
    SelectButtonModule,
    InputNumberModule,
    ButtonModule,
    MessageModule,
  ],
  templateUrl: './vsds-crosssection.component.html',
  styleUrl:    './vsds-crosssection.component.scss',
})
export class VsdsCrossSectionComponent implements OnInit, OnDestroy {

  @ViewChild('mapContainer')
  private mapContainerRef?: ElementRef<HTMLDivElement>;
  @ViewChild('svgOverlay')
  private svgRef?: ElementRef<SVGSVGElement>;
  @ViewChild('heatmapCanvas')
  private heatmapRef?: ElementRef<HTMLCanvasElement>;
  @ViewChild('threeCanvas')
  private threeCanvasRef?: ElementRef<HTMLCanvasElement>;

  // ── Source selection ─────────────────────────────────────────────────
  sourceOptions: SelectOption[] = [
    { label: 'Bundles',  value: 'bundles' },
    { label: 'Database', value: 'db' },
  ];
  selectedSource  = 'bundles';
  filteredBundles: Bundle[] = [];
  selectedBundle:  Bundle | null = null;

  // ── Parameters ───────────────────────────────────────────────────────
  corridorD  = 200;
  binWidth   = 100;
  renderMode: 'heatmap' | '3d' = 'heatmap';
  renderModeOptions: SelectOption[] = [
    { label: '2D Heatmap', value: 'heatmap' },
    { label: '3D Bars',    value: '3d' },
  ];

  // ── Map state ────────────────────────────────────────────────────────
  view:     'map' | 'section' = 'map';
  selState: SelectionState    = 'idle';
  pointA:   GalacticPoint | null = null;
  pointB:   GalacticPoint | null = null;
  dragging: 'a' | 'b' | null = null;
  svgViewBox = galaxyMapViewBox();

  // ── Coverage ──────────────────────────────────────────────────────────
  mapUrl        = '';
  coverageCells: CoverageCell[] = [];
  cellSizeLy    = 1500;
  private maxCoverageCount = 1;

  // ── Scale ─────────────────────────────────────────────────────────────
  scale:         RhoScale | null = null;
  scaleMaxInput  = 0;
  rhoUnit:       'kly3' | 'ly3' = 'kly3';
  unitOptions: SelectOption[] = [
    { label: 'sys/kLy³', value: 'kly3' },
    { label: 'sys/ly³',  value: 'ly3' },
  ];

  // ── DB source parameters ─────────────────────────────────────────────
  xzStep    = 50;
  yStep     = 20;
  dbLoading = false;
  dbError:  string | null = null;

  // ── Data ──────────────────────────────────────────────────────────────
  points:    BundleSurveyPoint[] = [];
  loading    = false;
  loadError: string | null = null;

  // ── Section view ──────────────────────────────────────────────────────
  filteredCount   = 0;
  abLengthDisplay = 0;

  private bundleBaseUrl: string | null = null;

  // ── Phase 4: 3D bar renderer ───────────────────────────────────────────
  private barsRenderer?: THREE.WebGLRenderer;
  private barsScene?:    THREE.Scene;
  private barsCamera?:   THREE.OrthographicCamera;
  private barsControls?: OrbitControls;
  private barsAnimId     = 0;
  private barsResizeObs?: ResizeObserver;

  constructor(
    public  authService:    AuthService,
    private bundlesService: BundlesService,
    private vsdsService:    VsdsService,
    private cdr:            ChangeDetectorRef,
  ) {}

  // ── Lifecycle ──────────────────────────────────────────────────────────

  ngOnInit(): void {
    this.bundleBaseUrl = this.bundlesService.bundleBaseUrl;
    if (!this.bundleBaseUrl) {
      this.bundlesService.getConfig().subscribe({
        next: cfg => {
          this.bundleBaseUrl = cfg.bundleBaseUrl;
          this.mapUrl = galaxyMapUrl(cfg.bundleBaseUrl);
        },
      });
    } else {
      this.mapUrl = galaxyMapUrl(this.bundleBaseUrl);
    }

    this.bundlesService.listBundles().subscribe({
      next: bundles => {
        this.filteredBundles = bundles.filter(b =>
          b.measurementtype === 'vsds' &&
          b.status          === 'ready' &&
          b.subtype         === 'surveypoints',
        );
      },
    });
  }

  ngOnDestroy(): void {
    this.destroyBarsRenderer();
  }

  // ── Bundle loading ────────────────────────────────────────────────────

  async onBundleChange(): Promise<void> {
    if (!this.selectedBundle || !this.bundleBaseUrl) return;
    this.loading       = true;
    this.loadError     = null;
    this.points        = [];
    this.scale         = null;
    this.coverageCells = [];
    this.selState      = 'idle';
    this.pointA        = null;
    this.pointB        = null;
    this.view          = 'map';
    try {
      const url  = this.bundleBaseUrl + this.selectedBundle.filename;
      const data = await fetchBundle<BundleSurveyPoint>(url);
      this.points = data;
      this.initScale(data);
      this.buildCoverageCells();
    } catch (err) {
      this.loadError = err instanceof Error
        ? err.message : 'Failed to load bundle';
    } finally {
      this.loading = false;
      this.cdr.detectChanges();
    }
  }

  // ── DB loading ────────────────────────────────────────────────────────

  loadFromDB(): void {
    this.dbLoading     = true;
    this.dbError       = null;
    this.points        = [];
    this.scale         = null;
    this.coverageCells = [];
    this.selState      = 'idle';
    this.pointA        = null;
    this.pointB        = null;
    this.view          = 'map';
    this.vsdsService.listSectors(this.xzStep, this.yStep)
      .subscribe({
        next: (voxels: SectorVoxel[]) => {
          this.points = voxels.map(v => ({
            sysname:     '',
            zsample:     (v.y_min + v.y_max) / 2,
            x:           v.gc_x + 25.21875,
            y:           (v.y_min + v.y_max) / 2 - 20.90625,
            z:           v.gc_z + 25899.96875,
            gc_x:        v.gc_x,
            gc_y:        (v.y_min + v.y_max) / 2,
            gc_z:        v.gc_z,
            syscount:    v.count,
            maxdistance: 0,
            rho:         v.rho_avg,
          }));
          this.initScale(this.points);
          this.buildCoverageCells();
          this.dbLoading = false;
          this.cdr.detectChanges();
        },
        error: (err: HttpErrorResponse) => {
          this.dbError   =
            err.error?.message ?? 'Failed to load sectors';
          this.dbLoading = false;
        },
      });
  }

  // ── Scale ─────────────────────────────────────────────────────────────

  private initScale(pts: BundleSurveyPoint[]): void {
    let dMin =  Infinity;
    let dMax = -Infinity;
    for (const p of pts) {
      if (p.rho > 0 && p.rho < dMin) dMin = p.rho;
      if (p.rho > dMax)              dMax = p.rho;
    }
    if (!isFinite(dMin)) dMin = 1e-6;
    if (!isFinite(dMax)) dMax = 1;
    this.scale = {
      dataMin: dMin, dataMax: dMax,
      scaleMin: dMin, scaleMax: dMax,
    };
    this.scaleMaxInput = this.toDisplay(dMax);
  }

  onScaleMaxChange(): void {
    if (!this.scale) return;
    const raw = this.fromDisplay(this.scaleMaxInput);
    if (raw <= 0) return;
    this.scale = { ...this.scale, scaleMax: raw };
    if (this.view === 'section') this.renderHeatmap();
  }

  resetScaleMax(): void {
    if (!this.scale) return;
    this.scale         = { ...this.scale, scaleMax: this.scale.dataMax };
    this.scaleMaxInput = this.toDisplay(this.scale.dataMax);
    if (this.view === 'section') this.renderHeatmap();
  }

  onUnitChange(): void {
    if (!this.scale) return;
    this.scaleMaxInput = this.toDisplay(this.scale.scaleMax);
  }

  // ── Coverage cells ────────────────────────────────────────────────────

  private buildCoverageCells(): void {
    const w =
      this.mapContainerRef?.nativeElement?.clientWidth || 800;
    const lyPerPx = 90000 / w;
    this.cellSizeLy = Math.max(
      20, Math.round(10 * lyPerPx / 20) * 20,
    );

    const cellMap = new Map<string, number>();
    for (const pt of this.points) {
      const cx =
        Math.floor(pt.gc_x / this.cellSizeLy) * this.cellSizeLy;
      const cz =
        Math.floor(pt.gc_z / this.cellSizeLy) * this.cellSizeLy;
      const key = `${cx}_${cz}`;
      cellMap.set(key, (cellMap.get(key) ?? 0) + 1);
    }

    let maxCount = 1;
    const cells: CoverageCell[] = [];
    for (const [key, count] of cellMap) {
      const sep = key.indexOf('_');
      cells.push({
        gc_x:  Number(key.slice(0, sep)),
        gc_z:  Number(key.slice(sep + 1)),
        count,
      });
      if (count > maxCount) maxCount = count;
    }
    this.coverageCells    = cells;
    this.maxCoverageCount = maxCount;
  }

  coverageFill(count: number): string {
    const t = Math.min(count / this.maxCoverageCount, 1);
    const g = Math.round(t * 150);
    const b = Math.round(t * 200);
    return `rgb(0,${g},${b})`;
  }

  // ── Map interaction ───────────────────────────────────────────────────

  get corridorPolygon(): string {
    if (!this.pointA || !this.pointB) return '';
    const ax = this.pointA.gc_x, az = this.pointA.gc_z;
    const bx = this.pointB.gc_x, bz = this.pointB.gc_z;
    const dx = bx - ax, dz = bz - az;
    const len = Math.hypot(dx, dz);
    if (len === 0) return '';
    const px = -dz / len * this.corridorD;
    const pz =  dx / len * this.corridorD;
    return [
      `${ax + px},${az + pz}`,
      `${bx + px},${bz + pz}`,
      `${bx - px},${bz - pz}`,
      `${ax - px},${az - pz}`,
    ].join(' ');
  }

  private svgToGalactic(event: MouseEvent): GalacticPoint {
    const svgEl = this.svgRef?.nativeElement;
    if (!svgEl) return { gc_x: 0, gc_z: 0 };
    const pt  = svgEl.createSVGPoint();
    pt.x      = event.clientX;
    pt.y      = event.clientY;
    const ctm = svgEl.getScreenCTM();
    if (!ctm) return { gc_x: 0, gc_z: 0 };
    const p = pt.matrixTransform(ctm.inverse());
    return { gc_x: p.x, gc_z: p.y };
  }

  onMapClick(event: MouseEvent): void {
    if (this.dragging) return;
    const pt = this.svgToGalactic(event);
    switch (this.selState) {
      case 'idle':
        this.pointA   = pt;
        this.selState = 'a_set';
        break;
      case 'a_set':
        this.pointB   = pt;
        this.selState = 'ready';
        break;
    }
  }

  startDrag(event: MouseEvent, handle: 'a' | 'b'): void {
    event.stopPropagation();
    this.dragging = handle;
    const onUp = () => {
      this.dragging = null;
      document.removeEventListener('mouseup', onUp);
    };
    document.addEventListener('mouseup', onUp);
  }

  onMapMouseMove(event: MouseEvent): void {
    if (!this.dragging) return;
    const pt = this.svgToGalactic(event);
    if (this.dragging === 'a') this.pointA = pt;
    else                       this.pointB = pt;
  }

  // ── Section view ──────────────────────────────────────────────────────

  drawSection(): void {
    this.destroyBarsRenderer();
    this.view = 'section';
    setTimeout(() => {
      if (this.renderMode === 'heatmap') {
        this.renderHeatmap();
      } else {
        this.renderBars();
      }
    });
  }

  backToMap(): void {
    this.destroyBarsRenderer();
    this.view = 'map';
  }

  // ── Binning helper ────────────────────────────────────────────────────

  private computeBins(): BinResult | null {
    if (!this.pointA || !this.pointB || !this.points.length) {
      return null;
    }

    const ax = this.pointA.gc_x, az = this.pointA.gc_z;
    const bx = this.pointB.gc_x, bz = this.pointB.gc_z;
    const abLength = Math.sqrt(
      (bx - ax) ** 2 + (bz - az) ** 2,
    );
    if (abLength === 0) return null;

    const unitAB = {
      x: (bx - ax) / abLength,
      z: (bz - az) / abLength,
    };

    const filtered = this.points.filter(p => {
      const perp = Math.abs(
        (p.gc_x - ax) * unitAB.z - (p.gc_z - az) * unitAB.x,
      );
      return perp <= this.corridorD;
    });
    if (!filtered.length) return null;

    const binMap = new Map<string, number[]>();
    for (const p of filtered) {
      const along =
        (p.gc_x - ax) * unitAB.x + (p.gc_z - az) * unitAB.z;
      const xBin  = Math.floor(along / this.binWidth);
      const key   = `${xBin}_${p.zsample}`;
      if (!binMap.has(key)) binMap.set(key, []);
      binMap.get(key)!.push(p.rho);
    }

    const cells: BinCell[] = [];
    let xBinMin = Infinity, xBinMax = -Infinity;
    const ySet  = new Set<number>();

    for (const [key, rhos] of binMap.entries()) {
      const under   = key.indexOf('_');
      const xBin    = parseInt(key.slice(0, under), 10);
      const zsample = parseFloat(key.slice(under + 1));
      cells.push({
        xBin,
        along:  xBin * this.binWidth + this.binWidth / 2,
        zsample,
        avgRho: rhos.reduce((s, r) => s + r, 0) / rhos.length,
      });
      if (xBin < xBinMin) xBinMin = xBin;
      if (xBin > xBinMax) xBinMax = xBin;
      ySet.add(zsample);
    }

    const yValues = [...ySet].sort((a, b) => a - b);
    return {
      cells, abLength, unitAB, yValues,
      xBinMin, xBinMax,
      filteredCount: filtered.length,
    };
  }

  // ── Canvas rendering ──────────────────────────────────────────────────

  private renderHeatmap(): void {
    const canvas = this.heatmapRef?.nativeElement;
    if (!canvas || !this.scale) return;

    const parent = canvas.parentElement;
    if (!parent) return;
    canvas.width  = parent.clientWidth;
    canvas.height = parent.clientHeight;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    const result = this.computeBins();

    if (!result) {
      ctx.fillStyle    = '#888';
      ctx.font         = '14px sans-serif';
      ctx.textAlign    = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(
        'No survey points within the corridor.',
        canvas.width / 2,
        canvas.height / 2,
      );
      return;
    }

    const { cells, yValues, xBinMin, xBinMax,
            filteredCount, abLength } = result;

    this.filteredCount   = filteredCount;
    this.abLengthDisplay = Math.round(abLength);

    const numX = xBinMax - xBinMin + 1;

    // Canvas layout
    const lp = 60, bp = 40, rp = 20, tp = 20;
    const pw = canvas.width  - lp - rp;
    const ph = canvas.height - tp - bp;
    const cw = pw / numX;
    const ch = ph / yValues.length;

    // Draw grid cells
    for (const cell of cells) {
      const t = rhoLogT(
        cell.avgRho,
        this.scale.scaleMin,
        this.scale.scaleMax,
      );
      const [r, g, b] = colorScale(VIRIDIS, t);
      ctx.fillStyle   = `rgb(${r},${g},${b})`;
      const xi    = cell.xBin - xBinMin;
      const yi    = yValues.indexOf(cell.zsample);
      const yFlip = yValues.length - 1 - yi;
      ctx.fillRect(lp + xi * cw, tp + yFlip * ch, cw, ch);
    }

    // Axes
    ctx.strokeStyle = '#666';
    ctx.lineWidth   = 1;
    ctx.fillStyle   = '#ccc';
    ctx.font        = '11px sans-serif';

    // X axis line
    ctx.beginPath();
    ctx.moveTo(lp, tp + ph);
    ctx.lineTo(lp + pw, tp + ph);
    ctx.stroke();

    // X ticks and labels
    const xStep = Math.max(1, Math.ceil(numX / 10));
    ctx.textAlign    = 'center';
    ctx.textBaseline = 'top';
    for (let i = 0; i <= numX; i += xStep) {
      const xLy = (xBinMin + i) * this.binWidth;
      const px  = lp + i * cw;
      ctx.beginPath();
      ctx.moveTo(px, tp + ph);
      ctx.lineTo(px, tp + ph + 4);
      ctx.stroke();
      ctx.fillText(`${Math.round(xLy)}`, px, tp + ph + 6);
    }

    // Y axis line
    ctx.beginPath();
    ctx.moveTo(lp, tp);
    ctx.lineTo(lp, tp + ph);
    ctx.stroke();

    // Y ticks and labels
    const yStep =
      Math.max(1, Math.ceil(yValues.length / 12));
    ctx.textAlign    = 'right';
    ctx.textBaseline = 'middle';
    for (let yi = 0; yi < yValues.length; yi += yStep) {
      const yFlip = yValues.length - 1 - yi;
      const py    = tp + (yFlip + 0.5) * ch;
      ctx.beginPath();
      ctx.moveTo(lp, py);
      ctx.lineTo(lp - 4, py);
      ctx.stroke();
      ctx.fillText(`${yValues[yi]}`, lp - 6, py);
    }

    // Axis title — X
    ctx.fillStyle    = '#aaa';
    ctx.font         = '12px sans-serif';
    ctx.textAlign    = 'center';
    ctx.textBaseline = 'bottom';
    ctx.fillText(
      'Distance from A (ly)', lp + pw / 2, canvas.height,
    );

    // Axis title — Y (rotated)
    ctx.save();
    ctx.translate(12, tp + ph / 2);
    ctx.rotate(-Math.PI / 2);
    ctx.textBaseline = 'top';
    ctx.fillText('Z-sample (ly)', 0, 0);
    ctx.restore();
  }

  // ── 3D bar renderer ───────────────────────────────────────────────────

  private destroyBarsRenderer(): void {
    if (this.barsAnimId) {
      cancelAnimationFrame(this.barsAnimId);
      this.barsAnimId = 0;
    }
    this.barsResizeObs?.disconnect();
    this.barsResizeObs = undefined;
    this.barsControls?.dispose();
    this.barsControls = undefined;
    this.barsRenderer?.dispose();
    this.barsRenderer = undefined;
    this.barsScene    = undefined;
    this.barsCamera   = undefined;
  }

  private initThreeBarsScene(
    canvas:   HTMLCanvasElement,
    cells:    BinCell[],
    yValues:  number[],
    abLength: number,
  ): void {
    this.destroyBarsRenderer();

    const W = canvas.clientWidth  || 800;
    const H = canvas.clientHeight || 600;
    canvas.width  = W;
    canvas.height = H;

    this.barsScene = new THREE.Scene();
    this.barsScene.background = new THREE.Color(0x0a0a0f);
    this.barsScene.add(
      new THREE.AmbientLight(0xffffff, 0.5),
    );
    const dir = new THREE.DirectionalLight(0xffffff, 0.8);
    dir.position.set(abLength / 2, 2000, 3000);
    this.barsScene.add(dir);

    const layerH = 50;
    for (const cell of cells) {
      if (cell.avgRho <= 0) continue;
      const color = rhoToHex(
        cell.avgRho,
        this.scale!.scaleMin,
        this.scale!.scaleMax,
      );
      const geo  = new THREE.BoxGeometry(
        this.binWidth, layerH, 50,
      );
      const mat  = new THREE.MeshPhongMaterial({ color });
      const mesh = new THREE.Mesh(geo, mat);
      mesh.position.set(cell.along, cell.zsample, 0);
      this.barsScene.add(mesh);
    }

    const xCenter = abLength / 2;
    const yMin    = yValues.length ? yValues[0] : 0;
    const yMax    =
      yValues.length ? yValues[yValues.length - 1] : 100;
    const yCenter = (yMin + yMax) / 2;
    const aspect  = W / H;

    const dataW = abLength * 1.15;
    const dataH = (yMax - yMin + 100) * 1.15;

    let fW = dataW, fH = dataH;
    if (fW / fH < aspect) fW = fH * aspect;
    else                   fH = fW / aspect;

    this.barsCamera = new THREE.OrthographicCamera(
      xCenter - fW / 2, xCenter + fW / 2,
      yCenter + fH / 2, yCenter - fH / 2,
      -100000, 100000,
    );
    this.barsCamera.position.set(xCenter, yCenter, 5000);
    this.barsCamera.lookAt(xCenter, yCenter, 0);

    this.barsRenderer = new THREE.WebGLRenderer({
      canvas, antialias: true,
    });
    this.barsRenderer.setPixelRatio(window.devicePixelRatio);
    this.barsRenderer.setSize(W, H, false);

    this.barsControls = new OrbitControls(
      this.barsCamera, canvas,
    );
    this.barsControls.target.set(xCenter, yCenter, 0);
    this.barsControls.enableDamping = true;
    this.barsControls.dampingFactor = 0.05;
    this.barsControls.update();

    const parent = canvas.parentElement!;
    this.barsResizeObs = new ResizeObserver(() => {
      const nW = parent.clientWidth;
      const nH = parent.clientHeight;
      if (nW === 0 || nH === 0) return;
      canvas.width  = nW;
      canvas.height = nH;
      this.barsRenderer!.setSize(nW, nH, false);
      const nAsp = nW / nH;
      let nfW = dataW, nfH = dataH;
      if (nfW / nfH < nAsp) nfW = nfH * nAsp;
      else                   nfH = nfW / nAsp;
      const cam = this.barsCamera!;
      cam.left   = xCenter - nfW / 2;
      cam.right  = xCenter + nfW / 2;
      cam.top    = yCenter + nfH / 2;
      cam.bottom = yCenter - nfH / 2;
      cam.updateProjectionMatrix();
    });
    this.barsResizeObs.observe(parent);

    const animate = () => {
      this.barsAnimId = requestAnimationFrame(animate);
      this.barsControls!.update();
      this.barsRenderer!.render(
        this.barsScene!, this.barsCamera!,
      );
    };
    animate();
  }

  private renderBars(): void {
    const result = this.computeBins();
    if (!result || !this.scale || !this.threeCanvasRef) return;
    const { cells, abLength, yValues, filteredCount } = result;
    this.filteredCount   = filteredCount;
    this.abLengthDisplay = Math.round(abLength);
    this.initThreeBarsScene(
      this.threeCanvasRef.nativeElement,
      cells, yValues, abLength,
    );
  }

  // ── Display helpers ───────────────────────────────────────────────────

  get unitLabel(): string {
    return this.rhoUnit === 'kly3' ? 'sys/kLy³' : 'sys/ly³';
  }

  toDisplay(rho: number): number {
    return this.rhoUnit === 'kly3' ? rho * 1000 : rho;
  }

  fromDisplay(v: number): number {
    return this.rhoUnit === 'kly3' ? v / 1000 : v;
  }

  formatRho(rho: number | null | undefined): string {
    if (rho == null) return '\u2014';
    return this.toDisplay(rho).toExponential(3);
  }
}
