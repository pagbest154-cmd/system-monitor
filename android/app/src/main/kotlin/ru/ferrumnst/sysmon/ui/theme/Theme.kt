package ru.ferrumnst.sysmon.ui.theme

import androidx.compose.material3.ColorScheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Typography
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

object SysMonColors {
    val Bg = Color(0xFF0F1419)
    val Surface = Color(0xFF1A2332)
    val Surface2 = Color(0xFF243044)
    val Border = Color(0xFF2D3A4F)
    val Text = Color(0xFFE8EDF4)
    val Muted = Color(0xFF8B9CB3)
    val Accent = Color(0xFF3B82F6)
    val Ok = Color(0xFF22C55E)
    val Warn = Color(0xFFF59E0B)
    val Critical = Color(0xFFEF4444)
    val ChartBlue = Color(0xFF3B82F6)
    val ChartGreen = Color(0xFF22C55E)
    val ChartOrange = Color(0xFFF59E0B)
    val ChartPurple = Color(0xFFA78BFA)
}

private val DarkColorScheme = darkColorScheme(
    primary = SysMonColors.Accent,
    onPrimary = Color.White,
    primaryContainer = Color(0xFF1E3A5F),
    onPrimaryContainer = Color(0xFFBFDBFE),
    secondary = SysMonColors.Ok,
    onSecondary = Color(0xFF052E16),
    tertiary = SysMonColors.Warn,
    background = SysMonColors.Bg,
    onBackground = SysMonColors.Text,
    surface = SysMonColors.Surface,
    onSurface = SysMonColors.Text,
    surfaceVariant = SysMonColors.Surface2,
    onSurfaceVariant = SysMonColors.Muted,
    outline = SysMonColors.Border,
    error = SysMonColors.Critical,
)

private val LightColorScheme = lightColorScheme(
    primary = SysMonColors.Accent,
    onPrimary = Color.White,
    primaryContainer = Color(0xFFDBEAFE),
    onPrimaryContainer = Color(0xFF1E3A5F),
    secondary = SysMonColors.Ok,
    onSecondary = Color.White,
    tertiary = SysMonColors.Warn,
    background = Color(0xFFF0F4F8),
    onBackground = Color(0xFF0F1419),
    surface = Color.White,
    onSurface = Color(0xFF0F1419),
    surfaceVariant = Color(0xFFE8EDF4),
    onSurfaceVariant = Color(0xFF5B6B82),
    outline = Color(0xFFD0DAE8),
    error = SysMonColors.Critical,
)

/** All typography slots use scheme colors — default M3 Typography keeps light-theme black on unset slots. */
private fun sysMonTypography(colorScheme: ColorScheme): Typography {
    val primary = colorScheme.onSurface
    val secondary = colorScheme.onSurfaceVariant
    return Typography(
        headlineLarge = TextStyle(
            fontWeight = FontWeight.SemiBold,
            fontSize = 32.sp,
            lineHeight = 40.sp,
            color = primary,
        ),
        headlineMedium = TextStyle(
            fontWeight = FontWeight.SemiBold,
            fontSize = 22.sp,
            lineHeight = 28.sp,
            color = primary,
        ),
        headlineSmall = TextStyle(
            fontWeight = FontWeight.SemiBold,
            fontSize = 24.sp,
            lineHeight = 32.sp,
            color = primary,
        ),
        titleLarge = TextStyle(
            fontWeight = FontWeight.SemiBold,
            fontSize = 20.sp,
            lineHeight = 26.sp,
            color = primary,
        ),
        titleMedium = TextStyle(
            fontWeight = FontWeight.Medium,
            fontSize = 16.sp,
            lineHeight = 22.sp,
            color = primary,
        ),
        titleSmall = TextStyle(
            fontWeight = FontWeight.Medium,
            fontSize = 14.sp,
            lineHeight = 20.sp,
            color = secondary,
        ),
        bodyLarge = TextStyle(
            fontSize = 16.sp,
            lineHeight = 22.sp,
            color = primary,
        ),
        bodyMedium = TextStyle(
            fontSize = 14.sp,
            lineHeight = 20.sp,
            color = primary,
        ),
        bodySmall = TextStyle(
            fontSize = 12.sp,
            lineHeight = 16.sp,
            color = secondary,
        ),
        labelLarge = TextStyle(
            fontWeight = FontWeight.Medium,
            fontSize = 14.sp,
            lineHeight = 20.sp,
            color = primary,
        ),
        labelMedium = TextStyle(
            fontWeight = FontWeight.Medium,
            fontSize = 12.sp,
            lineHeight = 16.sp,
            color = secondary,
        ),
        labelSmall = TextStyle(
            fontWeight = FontWeight.Medium,
            fontSize = 11.sp,
            lineHeight = 16.sp,
            color = secondary,
        ),
    )
}

@Composable
fun SysMonTheme(
    darkTheme: Boolean = true,
    content: @Composable () -> Unit,
) {
    val colorScheme = if (darkTheme) DarkColorScheme else LightColorScheme
    val typography = remember(colorScheme) { sysMonTypography(colorScheme) }
    MaterialTheme(
        colorScheme = colorScheme,
        typography = typography,
        content = content,
    )
}
