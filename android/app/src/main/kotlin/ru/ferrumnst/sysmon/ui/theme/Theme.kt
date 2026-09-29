package ru.ferrumnst.sysmon.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Typography
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
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

private val SysMonTypography = Typography(
    headlineMedium = TextStyle(
        fontWeight = FontWeight.SemiBold,
        fontSize = 22.sp,
        lineHeight = 28.sp,
    ),
    titleLarge = TextStyle(
        fontWeight = FontWeight.SemiBold,
        fontSize = 20.sp,
        lineHeight = 26.sp,
    ),
    titleMedium = TextStyle(
        fontWeight = FontWeight.Medium,
        fontSize = 16.sp,
        lineHeight = 22.sp,
    ),
    bodyLarge = TextStyle(
        fontSize = 16.sp,
        lineHeight = 22.sp,
    ),
    bodyMedium = TextStyle(
        fontSize = 14.sp,
        lineHeight = 20.sp,
    ),
    labelMedium = TextStyle(
        fontWeight = FontWeight.Medium,
        fontSize = 12.sp,
        lineHeight = 16.sp,
    ),
)

@Composable
fun SysMonTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit,
) {
    val colorScheme = if (darkTheme) DarkColorScheme else LightColorScheme
    MaterialTheme(
        colorScheme = colorScheme,
        typography = SysMonTypography,
        content = content,
    )
}
