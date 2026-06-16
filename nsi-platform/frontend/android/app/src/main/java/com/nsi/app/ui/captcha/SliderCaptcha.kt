package com.nsi.app.ui.captcha

import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectHorizontalDragGestures
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nsi.app.theme.AppColors
import kotlin.math.roundToInt

@Composable
fun SliderCaptcha(
    modifier: Modifier = Modifier,
    onVerified: () -> Unit,
) {
    var progress by remember { mutableStateOf(0f) }
    var verified by remember { mutableStateOf(false) }

    val knobSize = 48.dp
    val verifiedColor = AppColors.Highlight

    BoxWithConstraints(
        modifier = modifier
            .fillMaxWidth()
            .height(knobSize)
            .clip(RoundedCornerShape(knobSize / 2))
            .background(if (verified) verifiedColor else AppColors.StepInactive),
    ) {
        val maxOffsetPx = with(LocalDensity.current) { maxWidth.toPx() } -
            with(LocalDensity.current) { knobSize.toPx() }

        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            Text(
                if (verified) "验证成功" else "请拖动滑块验证",
                fontSize = 14.sp,
                fontWeight = FontWeight.SemiBold,
                color = if (verified) AppColors.White else AppColors.TextSecondary,
            )
        }

        Box(
            modifier = Modifier
                .offset { IntOffset((progress * maxOffsetPx).roundToInt(), 0) }
                .size(knobSize)
                .clip(RoundedCornerShape(knobSize / 2))
                .background(if (verified) AppColors.White else AppColors.Primary)
                .pointerInput(Unit) {
                    detectHorizontalDragGestures(
                        onDragEnd = {
                            if (!verified && progress < 0.95f) progress = 0f
                        },
                    ) { _, dragAmount ->
                        if (verified) return@detectHorizontalDragGestures
                        progress = (progress + dragAmount / maxOffsetPx).coerceIn(0f, 1f)
                        if (progress >= 0.95f) {
                            progress = 1f
                            verified = true
                            onVerified()
                        }
                    }
                },
            contentAlignment = Alignment.Center,
        ) {
            Text(
                "▶",
                color = if (verified) verifiedColor else AppColors.White,
                fontSize = 18.sp,
            )
        }
    }
}
