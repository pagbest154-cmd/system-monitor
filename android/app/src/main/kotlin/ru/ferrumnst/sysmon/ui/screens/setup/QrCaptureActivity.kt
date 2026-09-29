package ru.ferrumnst.sysmon.ui.screens.setup

import com.google.zxing.BarcodeFormat
import com.journeyapps.barcodescanner.CaptureActivity
import com.journeyapps.barcodescanner.DecoratedBarcodeView
import com.journeyapps.barcodescanner.DefaultDecoderFactory
import com.journeyapps.barcodescanner.ScanOptions

/** Portrait QR scanner (default [CaptureActivity] opens landscape with a 1D barcode line). */
class QrCaptureActivity : CaptureActivity() {
    override fun initializeContent(): DecoratedBarcodeView {
        val scanner = super.initializeContent()
        scanner.barcodeView.decoderFactory = DefaultDecoderFactory(listOf(BarcodeFormat.QR_CODE))
        return scanner
    }
}

fun hubQrScanOptions(): ScanOptions =
    ScanOptions().apply {
        setDesiredBarcodeFormats(ScanOptions.QR_CODE)
        setPrompt("Наведите на QR из настроек hub")
        setBeepEnabled(false)
        setOrientationLocked(true)
        setCaptureActivity(QrCaptureActivity::class.java)
    }
