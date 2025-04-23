import { Injectable } from '@angular/core';
import { Compress } from '../../wailsjs/go/main/App';

@Injectable({
  providedIn: 'root'
})

export class CompressService {
  async compress(file: string) {
    await Compress(file);
  }
}
