package com.nsi.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.lifecycle.viewmodel.compose.viewModel
import com.nsi.app.models.AppViewModel
import com.nsi.app.theme.NSITheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            NSITheme {
                val viewModel: AppViewModel = viewModel()
                NSIApp(viewModel)
            }
        }
    }
}
